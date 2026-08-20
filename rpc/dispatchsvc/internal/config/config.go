package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql MysqlConf     `yaml:"mysql" json:"mysql"`
	Redis cfg.RedisConf `yaml:"myredis" json:"myredis"`
	// EnableMockDispatch 允许在 GEO 查不到司机时回退 mock 候选，仅用于联调演示。
	EnableMockDispatch bool `yaml:"enableMockDispatch" json:"enableMockDispatch"`
	// DispatchTimeoutSeconds 派单超时阈值（秒），超过仍未接单的 Pending 记录视为超时并可重派。
	DispatchTimeoutSeconds int64 `yaml:"dispatchTimeoutSeconds" json:"dispatchTimeoutSeconds"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
