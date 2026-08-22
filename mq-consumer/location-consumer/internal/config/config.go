package config

import "XiaoLong-Ridy/common/config"

type Config struct {
	RedisConf config.RedisConf
	Log       LogConfig
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
