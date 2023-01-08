package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

const kafkaConnectionTimeout = 10 * time.Second

var (
	errNoStartOffsetFound = errors.New("no start offset found")
	errNoEndOffsetFound   = errors.New("no end offset found")
)

type State struct {
	Projects map[string]*Project `json:"projects"`
}

type Module struct {
	AppCtx context.Context

	state                 *State
	stateStorage          *state.Storage
	stateMutex            *sync.RWMutex
	clients               map[string]*kadm.Client
	topicConsumingClients map[string]*kgo.Client
	topicConsumingCancels map[string]context.CancelFunc
}

func NewModule(stateStorage *state.Storage) (*Module, error) {
	module := &Module{
		state: &State{
			Projects: make(map[string]*Project),
		},
		stateStorage:          stateStorage,
		clients:               make(map[string]*kadm.Client),
		topicConsumingClients: make(map[string]*kgo.Client),
		topicConsumingCancels: make(map[string]context.CancelFunc),
		stateMutex:            &sync.RWMutex{},
	}

	err := module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID] = &Project{
		ID:         projectID,
		CurrentTab: "overview",
		Address:    "0.0.0.0:9092",
		AuthMethod: "plaintext",
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveCurrentTab(projectID, currentTab string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentTab = currentTab

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Address = address

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthMethod(projectID, authMethod string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthMethod = authMethod

	if authMethod == "plaintext" {
		m.state.Projects[projectID].AuthUsername = ""
		m.state.Projects[projectID].AuthPassword = ""
	}

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthUsername(projectID, authUsername string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthUsername = authUsername

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) SaveAuthPassword(projectID, authPassword string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthPassword = authPassword

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

func (m *Module) Connect(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: kafkaConnectionTimeout}}

	options := []kgo.Opt{
		kgo.SeedBrokers(m.state.Projects[projectID].Address),
		kgo.Dialer(tlsDialer.DialContext),
	}

	if m.state.Projects[projectID].AuthMethod == "saslssl" {
		options = append(
			options,
			kgo.SASL(plain.Auth{
				User: m.state.Projects[projectID].AuthUsername,
				Pass: m.state.Projects[projectID].AuthPassword,
			}.AsMechanism()),
		)
	}

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("cannot establish kafka connection: %w", err)
	}

	ctx := context.Background()

	err = client.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to kafka: %w", err)
	}

	m.state.Projects[projectID].IsConnected = true
	m.clients[projectID] = kadm.NewClient(client)

	if err := m.saveState(); err != nil {
		return nil, err
	}

	return m.state, nil
}

// nolint: funlen
func (m *Module) Topics(projectID string) (*TabTopics, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	tabTopics := &TabTopics{
		IsConnected: m.state.Projects[projectID].IsConnected,
	}

	if !tabTopics.IsConnected {
		return tabTopics, nil
	}

	client := m.clients[projectID]

	ctx := context.Background()

	kafkaTopics, err := client.ListTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	startOffsets, err := client.ListStartOffsets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list start offsets: %w", err)
	}

	endOffsets, err := client.ListEndOffsets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list end offsets: %w", err)
	}

	tabTopics.Count = len(kafkaTopics)
	tabTopics.List = make([]*TabTopicsTopic, 0, len(kafkaTopics))

	for _, kafkaTopic := range kafkaTopics.Sorted() {
		var messageCount int64

		for _, partition := range kafkaTopic.Partitions {
			startOffset, ok := startOffsets.Lookup(kafkaTopic.Topic, partition.Partition)
			if !ok {
				return nil, errNoStartOffsetFound
			}

			endOffset, ok := endOffsets.Lookup(kafkaTopic.Topic, partition.Partition)
			if !ok {
				return nil, errNoEndOffsetFound
			}

			messageCount += endOffset.Offset - startOffset.Offset
		}

		tabTopics.List = append(
			tabTopics.List,
			&TabTopicsTopic{
				Name:           kafkaTopic.Topic,
				PartitionCount: len(kafkaTopic.Partitions),
				MessageCount:   messageCount,
			},
		)
	}

	return tabTopics, nil
}

func (m *Module) Brokers(projectID string) (*TabBrokers, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	tabBrokers := &TabBrokers{
		IsConnected: m.state.Projects[projectID].IsConnected,
	}

	if !tabBrokers.IsConnected {
		return tabBrokers, nil
	}

	client := m.clients[projectID]

	ctx := context.Background()

	kafkaBrokers, err := client.ListBrokers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list brokers: %w", err)
	}

	tabBrokers.Count = len(kafkaBrokers)
	tabBrokers.List = make([]*TabBrokersBroker, 0, len(kafkaBrokers))

	for _, kafkaBroker := range kafkaBrokers {
		broker := &TabBrokersBroker{
			ID:   int(kafkaBroker.NodeID),
			Host: kafkaBroker.Host,
			Port: int(kafkaBroker.Port),
		}

		if kafkaBroker.Rack != nil {
			broker.Rack = *kafkaBroker.Rack
		}

		tabBrokers.List = append(tabBrokers.List, broker)
	}

	return tabBrokers, nil
}

func (m *Module) Consumers(projectID string) (*TabConsumers, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	tabConsumers := &TabConsumers{
		IsConnected: m.state.Projects[projectID].IsConnected,
	}

	if !tabConsumers.IsConnected {
		return tabConsumers, nil
	}

	client := m.clients[projectID]

	ctx := context.Background()

	kafkaGroups, err := client.DescribeGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumers: %w", err)
	}

	tabConsumers.Count = len(kafkaGroups)
	tabConsumers.List = make([]*TabConsumersConsumer, 0, len(kafkaGroups))

	for _, kafkaGroup := range kafkaGroups {
		consumer := &TabConsumersConsumer{
			Name:  kafkaGroup.Group,
			State: kafkaGroup.State,
		}

		tabConsumers.List = append(tabConsumers.List, consumer)
	}

	return tabConsumers, nil
}

// nolint: funlen, cyclop
func (m *Module) StartTopicConsuming(projectID, topic string, hoursAgo int) (*TopicOutput, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: kafkaConnectionTimeout}}
	timeFrom := time.Now().UTC().Add(-time.Hour * time.Duration(hoursAgo)).UnixMilli()

	options := []kgo.Opt{
		kgo.SeedBrokers(m.state.Projects[projectID].Address),
		kgo.Dialer(tlsDialer.DialContext),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AfterMilli(timeFrom)),
	}

	if m.state.Projects[projectID].AuthMethod == "saslssl" {
		options = append(
			options,
			kgo.SASL(plain.Auth{
				User: m.state.Projects[projectID].AuthUsername,
				Pass: m.state.Projects[projectID].AuthPassword,
			}.AsMechanism()),
		)
	}

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("cannot establish kafka connection: %w", err)
	}

	m.topicConsumingClients[projectID] = client

	adminClient := kadm.NewClient(client)

	ctx := context.Background()

	kafkaTopics, err := adminClient.ListTopics(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	kafkaTopic := kafkaTopics[topic]

	startOffsets, err := adminClient.ListStartOffsets(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to list start offsets: %w", err)
	}

	endOffsets, err := adminClient.ListEndOffsets(ctx, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to list end offsets: %w", err)
	}

	output := &TopicOutput{
		Partitions: make([]*TopicPartition, 0, len(kafkaTopic.Partitions)),
	}

	partitionMap := make(map[int]*TopicPartition, len(kafkaTopic.Partitions))

	for _, partition := range kafkaTopic.Partitions {
		startOffset, ok := startOffsets.Lookup(kafkaTopic.Topic, partition.Partition)
		if !ok {
			return nil, errNoStartOffsetFound
		}

		endOffset, ok := endOffsets.Lookup(kafkaTopic.Topic, partition.Partition)
		if !ok {
			return nil, errNoEndOffsetFound
		}

		outputPartition := &TopicPartition{
			ID:               int(partition.Partition),
			OffsetTotalStart: startOffset.Offset,
			OffsetTotalEnd:   endOffset.Offset,
			OffsetCurrentEnd: endOffset.Offset,
		}

		output.CountTotal += endOffset.Offset - startOffset.Offset
		output.Partitions = append(output.Partitions, outputPartition)

		partitionMap[int(partition.Partition)] = outputPartition
	}

	go func() {
		for {
			ctx, cancelFunc := context.WithCancel(context.Background())
			m.topicConsumingCancels[projectID] = cancelFunc

			fetches := client.PollFetches(ctx)

			var isCanceled bool

			for _, err := range fetches.Errors() {
				if errors.Is(err.Err, context.Canceled) {
					isCanceled = true

					break
				}

				log.Fatal(err)
			}

			if isCanceled {
				break
			}

			for _, message := range fetches.Records() {
				outputMessage := &TopicMessage{
					Timestamp:   message.Timestamp.UTC(),
					PartitionID: int(message.Partition),
					Offset:      message.Offset,
					Key:         string(message.Key),
					Data:        string(message.Value),
				}

				runtime.EventsEmit(
					m.AppCtx,
					fmt.Sprintf("kafka_message_%s", projectID),
					outputMessage,
				)
			}
		}
	}()

	return output, nil
}

func (m *Module) StopTopicConsuming(projectID string) error {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	m.topicConsumingCancels[projectID]()
	m.topicConsumingCancels[projectID] = nil

	m.topicConsumingClients[projectID].Close()
	m.topicConsumingClients[projectID] = nil

	return nil
}

func (m *Module) State() (*State, error) {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	return m.state, nil
}

func (m *Module) Close() {
	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	for _, client := range m.clients {
		client.Close()
	}
}

func (m *Module) readOrInitializeState() error {
	isLoaded, err := m.stateStorage.Load("kafka", m.state)
	if err != nil {
		return fmt.Errorf("failed to load a state: %w", err)
	}

	if isLoaded {
		return nil
	}

	err = m.stateStorage.Save("kafka", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}

func (m *Module) saveState() error {
	state := &State{}

	err := copier.CopyWithOption(state, m.state, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return fmt.Errorf("failed to copy a kafka state: %w", err)
	}

	err = m.stateStorage.Save("kafka", m.state)
	if err != nil {
		return fmt.Errorf("failed to store a state: %w", err)
	}

	return nil
}
