// Package config 定义管理后台 RPC 服务的运行配置。
package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 adminsvc 的完整配置，包含 go-zero RPC、MySQL、Redis 和鉴权参数。
type Config struct {
	// RpcServerConf 以内联形式承载 Name、ListenOn 等 go-zero RPC 基础字段。
	zrpc.RpcServerConf `yaml:",inline"`
	// 以下字段显式声明 yaml tag，与 etc/admin.yaml 及本地启动临时配置的 key 大小写保持一致。
	// 缺少 yaml tag 时 yaml.v3 按字段名小写匹配，会导致整段配置被静默忽略。
	MySQL     MySQLConfig        `json:"MySQL" yaml:"MySQL"`
	Cache     RedisConfig        `json:"Cache" yaml:"Cache"`
	Session   AuthConfig         `json:"Session" yaml:"Session"`
	OrdersRPC zrpc.RpcClientConf `json:"OrdersRPC,optional" yaml:"OrdersRPC"`
	// DispatchRPC 供后台订单详情查询真实派单记录。
	DispatchRPC zrpc.RpcClientConf `json:"DispatchRPC,optional" yaml:"DispatchRPC"`
	// UsersRPC 供后台查询用户优惠券历史，数据读取仍由 usersvc 负责。
	UsersRPC   zrpc.RpcClientConf `json:"UsersRPC,optional" yaml:"UsersRPC"`
	DriversRPC zrpc.RpcClientConf `json:"DriversRPC,optional" yaml:"DriversRPC"`
	// LocationsRPC 供后台订单轨迹回放查询，轨迹点读取仍由 locationsvc 负责。
	LocationsRPC zrpc.RpcClientConf `json:"LocationsRPC,optional" yaml:"LocationsRPC"`
	PricesRPC    zrpc.RpcClientConf `json:"PricesRPC,optional" yaml:"PricesRPC"`
	// PushRPC 供后台司机冻结、审核和风控处置后通知司机端。
	PushRPC zrpc.RpcClientConf `json:"PushRPC,optional" yaml:"PushRPC"`
	// DisableDownstreamRPC 仅供本地最小服务集使用。启用后不创建未启动下游服务的 gRPC 客户端，
	// 避免本地开发时由连接重试占用大量内存；默认 false，线上行为保持不变。
	DisableDownstreamRPC bool                       `json:"DisableDownstreamRPC,optional" yaml:"DisableDownstreamRPC"`
	MenuRoles            map[int32][]MenuItemConfig `json:"MenuRoles,optional" yaml:"MenuRoles"`
}


// MySQLConfig 定义本服务访问业务数据库所需的数据源。
type MySQLConfig struct {
	DSN string `yaml:"DSN"`
}

// RedisConfig 定义本服务保存管理员会话时使用的 Redis 连接信息。
type RedisConfig struct {
	Host     string `yaml:"Host"`
	Password string `yaml:"Password"`
	DB       int    `yaml:"DB"`
}

// AuthConfig 定义后台 token 有效期和 Redis key 前缀。
type AuthConfig struct {
	SessionTTLHours int    `yaml:"SessionTTLHours"`
	TokenPrefix     string `yaml:"TokenPrefix"`
}

// MenuItemConfig 定义角色可见菜单及其子项。
// 菜单展示由配置控制；实际接口访问权限仍由各 RPC 的服务端角色校验决定。
type MenuItemConfig struct {
	Name     string           `json:"Name" yaml:"Name"`
	Path     string           `json:"Path" yaml:"Path"`
	Icon     string           `json:"Icon" yaml:"Icon"`
	Perm     string           `json:"Perm" yaml:"Perm"`
	Children []MenuItemConfig `json:"Children,optional" yaml:"Children,omitempty"`
}
