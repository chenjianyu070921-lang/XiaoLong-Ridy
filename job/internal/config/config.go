package config

import (
	cconfig "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Log LogConfig

	// Mysql MySQL 连接配置
	Mysql cconfig.MysqlConf `yaml:"mysql" json:"mysql"`
	// RedisConf Redis 连接配置
	RedisConf cconfig.RedisConf `yaml:"redis" json:"redis"`

	// OrderRPC ordersvc 客户端，超时未接单订单自动取消任务使用
	OrderRPC zrpc.RpcClientConf `yaml:"orderrpc" json:"orderrpc"`
	// TimeoutSeconds 订单超时秒数，0 时默认 300（5 分钟）
	TimeoutSeconds int64 `yaml:"timeoutseconds" json:"timeoutseconds"`
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
