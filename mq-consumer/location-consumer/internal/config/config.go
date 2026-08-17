package config

import (
	cfg "XiaoLong-Ridy/common/config"
)

type Config struct {
	Kafka KafkaConfig
	Redis cfg.RedisConf
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
