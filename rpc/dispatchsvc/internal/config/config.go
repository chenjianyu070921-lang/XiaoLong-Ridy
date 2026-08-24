package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql MysqlConf     `yaml:"mysql" json:"mysql"`
	Redis cfg.RedisConf `yaml:"myredis" json:"myredis"`
	Kafka cfg.KafkaConf `yaml:"kafka" json:"kafka"`
	// EnableMockDispatch 允许在 GEO 查不到司机时回退 mock 候选，仅用于联调演示。
	EnableMockDispatch bool `yaml:"enableMockDispatch" json:"enableMockDispatch"`
	// DispatchTimeoutSeconds 派单超时阈值（秒），超过仍未接单的 Pending 记录视为超时并可重派。
	DispatchTimeoutSeconds int64 `yaml:"dispatchTimeoutSeconds" json:"dispatchTimeoutSeconds"`
	// DefaultCityCode 默认城市编码，用于定位 GEO 桶（driver:geo:<cityCode>）。
	// 为空时回退为 "default"，需与 locationsvc 写入的城市键保持一致，否则按城市派单会查空（P1-M4-5）。
	DefaultCityCode string `yaml:"defaultCityCode" json:"defaultCityCode"`
	// OrderRPC 订单服务客户端配置：派单前复核订单状态，防止取消/超时后的竞态派单（P0-M4-1）。
	OrderRPC zrpc.RpcClientConf `yaml:"orderrpc" json:"orderrpc"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
