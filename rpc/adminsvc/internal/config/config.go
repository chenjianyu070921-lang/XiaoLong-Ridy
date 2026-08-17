// Package config 定义管理后台 RPC 服务的运行配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 adminsvc 的完整配置，包含 go-zero RPC、MySQL、Redis 和鉴权参数。
type Config struct {
	zrpc.RpcServerConf
	MySQL      MySQLConfig        `json:"MySQL"`
	Cache      RedisConfig        `json:"Cache"`
	Session    AuthConfig         `json:"Session"`
	OrdersRPC  zrpc.RpcClientConf `json:"OrdersRPC,optional"`
	DriversRPC zrpc.RpcClientConf `json:"DriversRPC,optional"`
}

// MySQLConfig 定义本服务访问业务数据库所需的数据源。
type MySQLConfig struct {
	DSN string
}

// RedisConfig 定义本服务保存管理员会话时使用的 Redis 连接信息。
type RedisConfig struct {
	Host     string
	Password string
	DB       int
}

// AuthConfig 定义后台 token 有效期和 Redis key 前缀。
type AuthConfig struct {
	SessionTTLHours int
	TokenPrefix     string
}
