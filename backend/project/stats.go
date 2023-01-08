package project

type Stats struct {
	GRPCProjectCount   int `json:"grpcProjectCount"`
	ThriftProjectCount int `json:"thriftProjectCount"`
	KafkaProjectCount  int `json:"kafkaProjectCount"`
}
