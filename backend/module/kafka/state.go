package kafka

type AuthMethod string

const (
	AuthMethodPlaintext = "plaintext"
	AuthMethodSASLSSL   = "sasl_ssl"
)

type Tab string

const (
	TabOverview  = "overview"
	TabBrokers   = "brokers"
	TabTopics    = "topics"
	TabConsumers = "consumers"
)

type TopicConsumingStrategy string

const (
	TopicConsumingStrategyTime           = "time"
	TopicConsumingStrategyOffsetSpecific = "offset_specific"
	TopicConsumingStrategyOffsetNewest   = "offset_newest"
	TopicConsumingStrategyOffsetOldest   = "offset_oldest"
)

type State struct {
	ID           string     `json:"id"`
	Address      string     `json:"address"`
	AuthMethod   AuthMethod `json:"authMethod"`
	AuthUsername string     `json:"authUsername"`
	AuthPassword string     `json:"authPassword"`
	IsConnected  bool       `json:"isConnected"`
	CurrentTab   Tab        `json:"currentTab"`
}

type TabTopicsData struct {
	IsConnected bool                  `json:"isConnected"`
	Count       int                   `json:"count"`
	List        []*TabTopicsDataTopic `json:"list"`
}

type TabTopicsDataTopic struct {
	Name           string `json:"name"`
	PartitionCount int    `json:"partitionCount"`
	MessageCount   int64  `json:"messageCount"`
}

type TabBrokersData struct {
	IsConnected bool                    `json:"isConnected"`
	Count       int                     `json:"count"`
	List        []*TabBrokersDataBroker `json:"list"`
}

type TabBrokersDataBroker struct {
	ID   int    `json:"id"`
	Rack string `json:"rack"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

type TabConsumersData struct {
	IsConnected bool                        `json:"isConnected"`
	Count       int                         `json:"count"`
	List        []*TabConsumersDataConsumer `json:"list"`
}

type TabConsumersDataConsumer struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type TopicOutput struct {
	TopicName     string            `json:"topicName"`
	StartFromTime string            `json:"startFromTime"`
	CountTotal    int64             `json:"countTotal"`
	CountCurrent  int64             `json:"countCurrent"`
	Partitions    []*TopicPartition `json:"partitions"`
}

type TopicConsumingOutput struct {
	Messages []*TopicMessage `json:"messages"`
}

type TopicMessage struct {
	TimestampUnix      int64  `json:"timestampUnix"`
	TimestampFormatted string `json:"timestampFormatted"`
	PartitionID        int    `json:"partitionID"`
	Offset             int64  `json:"offset"`
	Key                string `json:"key"`
	Data               string `json:"data"`
	Headers            string `json:"headers"`
}

type TopicPartition struct {
	ID                 int   `json:"id"`
	OffsetTotalStart   int64 `json:"offsetTotalStart"`
	OffsetTotalEnd     int64 `json:"offsetTotalEnd"`
	OffsetCurrentStart int64 `json:"offsetCurrentStart"`
	OffsetCurrentEnd   int64 `json:"offsetCurrentEnd"`
}
