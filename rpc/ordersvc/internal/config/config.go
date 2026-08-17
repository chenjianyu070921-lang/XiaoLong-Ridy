package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql       MysqlConf         `yaml:"mysql" json:"mysql"`
	Kafka       KafkaConf         `yaml:"kafka" json:"kafka"`
	DispatchRPC zrpc.RpcClientConf `yaml:"dispatchrpc" json:"dispatchrpc"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

type KafkaConf struct {
	Brokers []string `yaml:"brokers" json:"brokers"`
	Topic   string   `yaml:"topic" json:"topic"`
	Group   string   `yaml:"group" json:"group"`
}
