// Package config 定义管理后台 RPC 服务的运行配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 adminsvc 的完整配置，包含 go-zero RPC、MySQL、Redis 和鉴权参数。
type Config struct {
	// RpcServerConf 以内联形式承载 Name、ListenOn 等 go-zero RPC 基础字段。
	zrpc.RpcServerConf `yaml:",inline"`
	MySQL              MySQLConfig        `json:"MySQL"`
	Cache              RedisConfig        `json:"Cache"`
	Session            AuthConfig         `json:"Session"`
	OrdersRPC          zrpc.RpcClientConf `json:"OrdersRPC,optional"`
	DriversRPC         zrpc.RpcClientConf `json:"DriversRPC,optional"`
	PricesRPC          zrpc.RpcClientConf `json:"PricesRPC,optional"`
	// DisableDownstreamRPC 仅供本地最小服务集使用。启用后不创建未启动下游服务的 gRPC 客户端，
	// 避免本地开发时由连接重试占用大量内存；默认 false，线上行为保持不变。
	DisableDownstreamRPC bool                       `json:"DisableDownstreamRPC,optional" yaml:"DisableDownstreamRPC"`
	MenuRoles            map[int32][]MenuItemConfig `json:"MenuRoles,optional"`
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

// MenuItemConfig 定义角色可见菜单及其子项。
// 菜单展示由配置控制；实际接口访问权限仍由各 RPC 的服务端角色校验决定。
type MenuItemConfig struct {
	Name     string           `json:"Name"`
	Path     string           `json:"Path"`
	Icon     string           `json:"Icon"`
	Perm     string           `json:"Perm"`
	Children []MenuItemConfig `json:"Children,optional"`
}
