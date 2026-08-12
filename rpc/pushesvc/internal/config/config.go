package config

import (
	commonconfig "XiaoLong-Ridy/common/config"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql     commonconfig.MysqlConf
	RedisConf commonconfig.RedisConf
	SMS       SMSConfig
	Push      PushConfig
}

type SMSConfig struct {
	Provider  string
	AccessKey string
	SecretKey string
	SignName  string
}

type PushConfig struct {
	Provider     string
	AppKey       string
	MasterSecret string
}
