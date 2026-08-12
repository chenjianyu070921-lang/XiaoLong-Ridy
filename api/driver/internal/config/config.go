// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	// 下游 RPC 依赖（待 rpc 服务生成后在 svc 中初始化客户端）
	UserRpc     zrpc.RpcClientConf
	OrderRpc    zrpc.RpcClientConf
	LocationRpc zrpc.RpcClientConf
	PayRpc      zrpc.RpcClientConf
	PushRpc     zrpc.RpcClientConf
}
