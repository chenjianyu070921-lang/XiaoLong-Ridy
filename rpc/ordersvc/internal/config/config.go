package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql MysqlConf `yaml:"mysql"`
	Redis RedisConf `yaml:"redis"`
	Kafka KafkaConf `yaml:"kafka"`
}

type MysqlConf struct {
	DSN string `yaml:"DSN"`
}

type RedisConf struct {
	Addr string `yaml:"Addr"`
}

type KafkaConf struct {
	Brokers []string `yaml:"Brokers"`
	Topic   string   `yaml:"Topic"`
	Group   string   `yaml:"Group"`
}
