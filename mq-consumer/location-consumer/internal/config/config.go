package config

import (
	cfg "XiaoLong-Ridy/common/config"
)

type Config struct {
	Kafka KafkaConfig
	Redis cfg.RedisConf
	Log   LogConfig
	// DefaultCityCode 位置写入默认城市编码，需与 dispatchsvc 保持一致。
	DefaultCityCode string `yaml:"defaultCityCode" json:"defaultCityCode"`
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
