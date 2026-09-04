package config

import (
	"time"

	"XiaoLong-Ridy/common/alipay"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	HttpAddr   string        `yaml:"httpAddr" json:"httpAddr"`
	Mysql      MysqlConf     `yaml:"mysql" json:"mysql"`
	Alipay     alipay.Config `yaml:"alipay" json:"alipay"`
	Kafka      KafkaConf     `yaml:"kafka" json:"kafka"`
	Ordersvc   OrdersvcConf  `yaml:"ordersvc" json:"ordersvc"`
	Reconcile  ReconcileConf `yaml:"reconcile" json:"reconcile"`
	AutoSettle AutoSettleConf `yaml:"autoSettle" json:"autoSettle"`
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

// ReconcileConf 支付渠道对账任务配置。
type ReconcileConf struct {
	Interval string `yaml:"interval" json:"interval"` // 如 "5m"
	Lookback string `yaml:"lookback" json:"lookback"` // 如 "30m"
}

func (r ReconcileConf) IntervalDuration() time.Duration {
	if d, err := time.ParseDuration(r.Interval); err == nil {
		return d
	}
	return 5 * time.Minute
}

func (r ReconcileConf) LookbackDuration() time.Duration {
	if d, err := time.ParseDuration(r.Lookback); err == nil {
		return d
	}
	return 30 * time.Minute
}

// AutoSettleConf 自动结算任务配置。
type AutoSettleConf struct {
	Interval              string  `yaml:"interval" json:"interval"`
	DefaultCommissionRate float64 `yaml:"defaultCommissionRate" json:"defaultCommissionRate"`
}

func (a AutoSettleConf) IntervalDuration() time.Duration {
	if d, err := time.ParseDuration(a.Interval); err == nil {
		return d
	}
	return 10 * time.Minute
}
