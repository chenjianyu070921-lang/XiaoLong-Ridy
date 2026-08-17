package config

import (
	commonconfig "XiaoLong-Ridy/common/config"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Log       LogConfig
	Mysql     commonconfig.MysqlConf
	RedisConf commonconfig.RedisConf
	OrderRPC  zrpc.RpcClientConf
}

type LogConfig struct {
	ServiceName string `json:"serviceName"`
	Mode        string `json:"mode"`
	Level       string `json:"level"`
}
