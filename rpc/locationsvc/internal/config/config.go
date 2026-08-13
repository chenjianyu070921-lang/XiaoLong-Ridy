package config

import (
	commonconfig "XiaoLong-Ridy/common/config"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql      commonconfig.MysqlConf
	RedisConf  commonconfig.RedisConf
	MapService MapServiceConfig
}

type MapServiceConfig struct {
	ApiKey   string
	Provider string
	BaseUrl  string
}
