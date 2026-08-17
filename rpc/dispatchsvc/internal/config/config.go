package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql MysqlConf `yaml:"mysql" json:"mysql"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
