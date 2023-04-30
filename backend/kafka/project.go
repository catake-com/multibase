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

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/multibase-io/multibase/backend/pkg/state"
)

const (
	kafkaConnectionTimeout = 10 * time.Second

	consumingTimeFromLayout = "2006-01-02 15:04:05 Z07:00"
)

var (
	errNoStartOffsetFound = errors.New("no start offset found")
	errNoEndOffsetFound   = errors.New("no end offset found")
)

type Project struct {
	state                *State
	stateMutex           sync.RWMutex
	stateStorage         *state.Storage
	client               *kadm.Client
	topicConsumingClient *kgo.Client
	topicConsumingCancel context.CancelFunc
}

func NewProject(projectID string, stateStorage *state.Storage) (*Project, error) {
	project := &Project{
		state: &State{
			ID:         projectID,
			CurrentTab: TabOverview,
			Address:    "0.0.0.0:9092",
			AuthMethod: AuthMethodPlaintext,
		},
		stateStorage: stateStorage,
	}

	if err := project.saveState(); err != nil {
		return nil, err
	}

	return project, nil
}

func (p *Project) SaveState(state *State) error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	if state.AuthMethod == AuthMethodPlaintext {
		state.AuthUsername = ""
		state.AuthPassword = ""
	}

	p.state = state

	return p.saveState()
}

func (p *Project) Connect() error {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: kafkaConnectionTimeout}}

	options := []kgo.Opt{
		kgo.SeedBrokers(p.state.Address),
		kgo.Dialer(tlsDialer.DialContext),
	}

	if p.state.AuthMethod == AuthMethodSASLSSL {
		options = append(
			options,
			kgo.SASL(plain.Auth{
				User: p.state.AuthUsername,
				Pass: p.state.AuthPassword,
			}.AsMechanism()),
		)
	}

	client, err := kgo.NewClient(options...)
	if err != nil {
		return fmt.Errorf("cannot establish kafka connection: %w", err)
	}

	ctx := context.Background()

	err = client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect to kafka: %w", err)
	}

	p.state.IsConnected = true
	p.client = kadm.NewClient(client)

	return p.saveState()
}

// nolint: funlen
func (p *Project) Topics() (*TabTopicsData, error) {
	tabTopicsData := &TabTopicsData{
		IsConnected: p.state.IsConnected,
	}

	if !tabTopicsData.IsConnected {
		return tabTopicsData, nil
	}

	ctx := context.Background()

	kafkaTopics, err := p.client.ListTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	startOffsets, err := p.client.ListStartOffsets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list start offsets: %w", err)
	}

	endOffsets, err := p.client.ListEndOffsets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list end offsets: %w", err)
	}

	tabTopicsData.Count = len(kafkaTopics)
	tabTopicsData.List = make([]*TabTopicsDataTopic, 0, len(kafkaTopics))

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

		tabTopicsData.List = append(
			tabTopicsData.List,
			&TabTopicsDataTopic{
				Name:           kafkaTopic.Topic,
				PartitionCount: len(kafkaTopic.Partitions),
				MessageCount:   messageCount,
			},
		)
	}

	return tabTopicsData, nil
}

func (p *Project) Brokers() (*TabBrokersData, error) {
	tabBrokersData := &TabBrokersData{
		IsConnected: p.state.IsConnected,
	}

	if !tabBrokersData.IsConnected {
		return tabBrokersData, nil
	}

	ctx := context.Background()

	kafkaBrokers, err := p.client.ListBrokers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list brokers: %w", err)
	}

	tabBrokersData.Count = len(kafkaBrokers)
	tabBrokersData.List = make([]*TabBrokersDataBroker, 0, len(kafkaBrokers))

	for _, kafkaBroker := range kafkaBrokers {
		broker := &TabBrokersDataBroker{
			ID:   int(kafkaBroker.NodeID),
			Host: kafkaBroker.Host,
			Port: int(kafkaBroker.Port),
		}

		if kafkaBroker.Rack != nil {
			broker.Rack = *kafkaBroker.Rack
		}

		tabBrokersData.List = append(tabBrokersData.List, broker)
	}

	return tabBrokersData, nil
}

func (p *Project) Consumers() (*TabConsumersData, error) {
	tabConsumersData := &TabConsumersData{
		IsConnected: p.state.IsConnected,
	}

	if !tabConsumersData.IsConnected {
		return tabConsumersData, nil
	}

	ctx := context.Background()

	kafkaGroups, err := p.client.DescribeGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumers: %w", err)
	}

	tabConsumersData.Count = len(kafkaGroups)
	tabConsumersData.List = make([]*TabConsumersDataConsumer, 0, len(kafkaGroups))

	for _, kafkaGroup := range kafkaGroups {
		consumer := &TabConsumersDataConsumer{
			Name:  kafkaGroup.Group,
			State: kafkaGroup.State,
		}

		tabConsumersData.List = append(tabConsumersData.List, consumer)
	}

	return tabConsumersData, nil
}

// nolint: funlen, cyclop
func (p *Project) StartTopicConsuming(ctx context.Context, topic, timeFrom string) (*TopicOutput, error) {
	p.stateMutex.Lock()
	defer p.stateMutex.Unlock()

	tlsDialer := &tls.Dialer{NetDialer: &net.Dialer{Timeout: kafkaConnectionTimeout}}

	timeFromParsed, err := time.Parse(consumingTimeFromLayout, timeFrom)
	if err != nil {
		return nil, fmt.Errorf("cannot parse kafka consuming time from: %w", err)
	}

	options := []kgo.Opt{
		kgo.SeedBrokers(p.state.Address),
		kgo.Dialer(tlsDialer.DialContext),
		kgo.ConsumeTopics(topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AfterMilli(timeFromParsed.UnixMilli())),
	}

	if p.state.AuthMethod == AuthMethodSASLSSL {
		options = append(
			options,
			kgo.SASL(plain.Auth{
				User: p.state.AuthUsername,
				Pass: p.state.AuthPassword,
			}.AsMechanism()),
		)
	}

	client, err := kgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("cannot establish kafka connection: %w", err)
	}

	p.topicConsumingClient = client

	adminClient := kadm.NewClient(client)

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
		TopicName:     topic,
		StartFromTime: timeFrom,
		Partitions:    make([]*TopicPartition, 0, len(kafkaTopic.Partitions)),
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
			ctx, cancelFunc := context.WithCancel(ctx)
			p.topicConsumingCancel = cancelFunc

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
					ctx,
					fmt.Sprintf("kafka_message_%s", p.state.ID),
					outputMessage,
				)
			}
		}
	}()

	return output, nil
}

func (p *Project) StopTopicConsuming() error {
	p.stateMutex.RLock()
	defer p.stateMutex.RUnlock()

	p.topicConsumingCancel()
	p.topicConsumingCancel = nil

	p.topicConsumingClient.Close()
	p.topicConsumingClient = nil

	return nil
}

func (p *Project) Close() error {
	if p.topicConsumingCancel != nil {
		p.topicConsumingCancel()
	}

	if p.topicConsumingClient != nil {
		p.topicConsumingClient.Close()
	}

	if p.client != nil {
		p.client.Close()
	}

	return nil
}

func (p *Project) saveState() error {
	copiedState := *p.state
	copiedState.IsConnected = false
	copiedState.CurrentTab = ""

	err := p.stateStorage.Save(p.state.ID, &copiedState)
	if err != nil {
		return fmt.Errorf("failed to store a kafka project: %w", err)
	}

	return nil
}
