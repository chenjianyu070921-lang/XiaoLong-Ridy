package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql MysqlConf     `yaml:"mysql" json:"mysql"`
	Redis cfg.RedisConf `yaml:"myredis" json:"myredis"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
