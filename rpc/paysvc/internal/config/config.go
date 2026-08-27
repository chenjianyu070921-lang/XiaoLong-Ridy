package config

import (
	"XiaoLong-Ridy/common/alipay"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	HttpAddr string        `yaml:"httpAddr" json:"httpAddr"`
	Mysql    MysqlConf     `yaml:"mysql" json:"mysql"`
	Alipay   alipay.Config `yaml:"alipay" json:"alipay"`
	Kafka    KafkaConf     `yaml:"kafka" json:"kafka"`
	Ordersvc OrdersvcConf  `yaml:"ordersvc" json:"ordersvc"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

// KafkaConf Kafka 生产者配置。
type KafkaConf struct {
	Brokers []string `yaml:"brokers" json:"brokers"`
	Topic   string   `yaml:"topic" json:"topic"`
}

// OrdersvcConf 订单服务 RPC 客户端配置（直连）。
type OrdersvcConf struct {
	Target string `yaml:"target" json:"target"` // 如 127.0.0.1:50051
}
