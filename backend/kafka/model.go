package kafka

import (
	"time"
)

type TabTopics struct {
	IsConnected bool              `json:"isConnected"`
	Count       int               `json:"count"`
	List        []*TabTopicsTopic `json:"list"`
}

type TabTopicsTopic struct {
	Name           string `json:"name"`
	PartitionCount int    `json:"partitionCount"`
	MessageCount   int64  `json:"messageCount"`
}

type TabBrokers struct {
	IsConnected bool                `json:"isConnected"`
	Count       int                 `json:"count"`
	List        []*TabBrokersBroker `json:"list"`
}

type TabBrokersBroker struct {
	ID   int    `json:"id"`
	Rack string `json:"rack"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type TabConsumers struct {
	IsConnected bool                    `json:"isConnected"`
	Count       int                     `json:"count"`
	List        []*TabConsumersConsumer `json:"list"`
}

type TabConsumersConsumer struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type TopicOutput struct {
	CountTotal   int64             `json:"countTotal"`
	CountCurrent int64             `json:"countCurrent"`
	Partitions   []*TopicPartition `json:"partitions"`
}

type TopicMessage struct {
	Timestamp   time.Time         `json:"timestamp"`
	PartitionID int               `json:"partitionID"`
	Offset      int64             `json:"offset"`
	Key         string            `json:"key"`
	Data        string            `json:"data"`
	Headers     map[string]string `json:"headers"`
}

type TopicPartition struct {
	ID                 int   `json:"id"`
	OffsetTotalStart   int64 `json:"offsetTotalStart"`
	OffsetTotalEnd     int64 `json:"offsetTotalEnd"`
	OffsetCurrentStart int64 `json:"offsetCurrentStart"`
	OffsetCurrentEnd   int64 `json:"offsetCurrentEnd"`
}
