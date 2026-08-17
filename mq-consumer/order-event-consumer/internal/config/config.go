package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Redis       cfg.RedisConf      `yaml:"redis" json:"redis"`
	DispatchRPC zrpc.RpcClientConf `yaml:"dispatchrpc" json:"dispatchrpc"`
	Log         LogConfig          `yaml:"log" json:"log"`
}

type LogConfig struct {
	ServiceName string `yaml:"serviceName" json:"serviceName"`
	Mode        string `yaml:"mode" json:"mode"`
	Level       string `yaml:"level" json:"level"`
}
