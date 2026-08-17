package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Log LogConfig

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
