package config

type Config struct {
	Kafka KafkaConfig
	Redis RedisConfig
	Log   LogConfig
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Group   string
}

type RedisConfig struct {
	Host string
	Pass string
	Type string
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
