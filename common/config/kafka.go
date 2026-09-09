package config

// KafkaConf 通用 Kafka 配置结构体，所有服务直接嵌入。
// 与支付模块（paysvc）的 KafkaConf 形态对齐：Brokers 指定 broker 地址列表。
type KafkaConf struct {
	Brokers []string `json:"brokers" yaml:"brokers"`
}
