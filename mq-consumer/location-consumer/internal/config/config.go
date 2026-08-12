package config

type Config struct {
	Kafka KafkaConfig
	Log   LogConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Group   string
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
