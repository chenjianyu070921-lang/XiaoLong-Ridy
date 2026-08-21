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

	// DispatchRPC dispatchsvc 客户端，派单超时重派任务使用
	DispatchRPC zrpc.RpcClientConf `yaml:"dispatchrpc" json:"dispatchrpc"`
	// DispatchTimeoutSeconds 派单超时秒数，0 时默认 60（1 分钟）
	DispatchTimeoutSeconds int64 `yaml:"dispatchtimeoutseconds" json:"dispatchtimeoutseconds"`
}

type LogConfig struct {
	ServiceName string
	Mode        string
	Level       string
}
