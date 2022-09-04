package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/99designs/keyring"
	"github.com/adrg/xdg"
	"github.com/jinzhu/copier"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/multierr"

	"github.com/multibase-io/multibase/backend/pkg/storage"
)

const kafkaConnectionTimeout = 10 * time.Second

var (
	errNoStartOffsetFound = errors.New("no start offset found")
	errNoEndOffsetFound   = errors.New("no end offset found")
)

type State struct {
	Projects map[string]*StateProject `json:"projects"`
}

type StateProject struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	AuthMethod   string `json:"authMethod"`
	AuthUsername string `json:"authUsername" copier:"-"`
	AuthPassword string `json:"authPassword" copier:"-"`
	IsConnected  bool   `json:"isConnected" copier:"-"`
	CurrentTab   string `json:"currentTab" copier:"-"`
}

type Module struct {
	AppCtx                context.Context
	configFilePath        string
	ring                  keyring.Keyring
	state                 *State
	stateMutex            *sync.RWMutex
	stateTimer            *time.Timer
	clients               map[string]*kadm.Client
	topicConsumingClients map[string]*kgo.Client
	topicConsumingCancels map[string]context.CancelFunc
}

func NewModule() (*Module, error) {
	configFilePath, err := xdg.ConfigFile("multibase/kafka")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve kafka config path: %w", err)
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName:              "multibase_kafka",
		KeychainTrustApplication: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open kafka keyring: %w", err)
	}

	module := &Module{
		configFilePath:        configFilePath,
		ring:                  ring,
		clients:               make(map[string]*kadm.Client),
		topicConsumingClients: make(map[string]*kgo.Client),
		topicConsumingCancels: make(map[string]context.CancelFunc),
		state: &State{
			Projects: make(map[string]*StateProject),
		},
		stateMutex: &sync.RWMutex{},
	}

	err = module.readOrInitializeState()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (m *Module) CreateNewProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID] = &StateProject{
		ID:         projectID,
		CurrentTab: "overview",
		Address:    "0.0.0.0:9092",
		AuthMethod: "plaintext",
	}

	if err := m.saveStateToFile(); err != nil {
		return nil, fmt.Errorf("failed to save state to file: %w", err)
	}

	return m.state, nil
}

func (m *Module) DeleteProject(projectID string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	delete(m.state.Projects, projectID)

	passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
	_ = m.ring.Remove(passwordKey)

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveCurrentTab(projectID, currentTab string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].CurrentTab = currentTab

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAddress(projectID, address string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].Address = address

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAuthMethod(projectID, authMethod string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthMethod = authMethod

	if authMethod == "plaintext" {
		passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)
		usernameKey := fmt.Sprintf("%s_AuthUsername", projectID)
		_ = m.ring.Remove(passwordKey)
		_ = m.ring.Remove(usernameKey)

		m.state.Projects[projectID].AuthUsername = ""
		m.state.Projects[projectID].AuthPassword = ""
	}

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAuthUsername(projectID, authUsername string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthUsername = authUsername

	usernameKey := fmt.Sprintf("%s_AuthUsername", projectID)

	if authUsername == "" {
		_ = m.ring.Remove(usernameKey)
	} else {
		err := m.ring.Set(keyring.Item{
			Key:  usernameKey,
			Data: []byte(authUsername),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set a kafka keyring key: %w", err)
		}
	}

	m.saveState()

	return m.state, nil
}

func (m *Module) SaveAuthPassword(projectID, authPassword string) (*State, error) {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	m.state.Projects[projectID].AuthPassword = authPassword

	passwordKey := fmt.Sprintf("%s_AuthPassword", projectID)

	if authPassword == "" {
		_ = m.ring.Remove(passwordKey)
	} else {
		err := m.ring.Set(keyring.Item{
			Key:  passwordKey,
			Data: []byte(authPassword),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set a kafka keyring key: %w", err)
		}
	}

	m.saveState()

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

	m.saveState()

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
	_, err := os.Stat(m.configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to describe a kafka config file: %w", err)
		}

		return m.initializeState()
	}

	return m.readState()
}

func (m *Module) initializeState() (rerr error) {
	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	data, err := json.Marshal(m.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedState, err := storage.Encrypt(storage.DefaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	_, err = file.Write(encryptedState)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

// nolint: cyclop
func (m *Module) readState() (rerr error) {
	file, err := os.Open(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to open a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read state from file: %w", err)
	}

	decryptedData, err := storage.Decrypt(storage.DefaultPassword, data)
	if err != nil {
		if errors.Is(err, storage.ErrNoData) {
			return nil
		}

		return fmt.Errorf("failed to decrypt state: %w", err)
	}

	err = json.Unmarshal(decryptedData, m.state)
	if err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	for _, project := range m.state.Projects {
		project.CurrentTab = "overview"

		if project.AuthMethod == "plaintext" {
			continue
		}

		authPassword, err := m.ring.Get(fmt.Sprintf("%s_AuthPassword", project.ID))
		if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("failed to get a kafka keyring key: %w", err)
		}

		project.AuthPassword = string(authPassword.Data)

		authUsername, err := m.ring.Get(fmt.Sprintf("%s_AuthUsername", project.ID))
		if err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("failed to get a kafka keyring key: %w", err)
		}

		project.AuthUsername = string(authUsername.Data)
	}

	return nil
}

func (m *Module) saveState() {
	if m.stateTimer != nil {
		_ = m.stateTimer.Stop()
	}

	m.stateTimer = time.AfterFunc(storage.DefaultStatePersistenceDelay, func() {
		err := m.saveStateToFile()
		if err != nil {
			log.Println(fmt.Errorf("failed to save state to a file: %w", err))
		}
	})
}

func (m *Module) saveStateToFile() (rerr error) {
	state := &State{}

	err := copier.CopyWithOption(state, m.state, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		return fmt.Errorf("failed to copy a kafka state: %w", err)
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	encryptedData, err := storage.Encrypt(storage.DefaultPassword, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt state: %w", err)
	}

	file, err := os.Create(m.configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create/truncate a kafka config file: %w", err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			rerr = multierr.Combine(rerr, fmt.Errorf("failed to close a config file: %w", err))
		}
	}()

	_, err = file.Write(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}
