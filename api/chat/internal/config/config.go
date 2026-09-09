package config

import "github.com/zeromicro/go-zero/zrpc"

// Config maps api/chat/etc/chat.yaml。
// 注意：go-zero 配置解析只认 json 标签（忽略 yaml 标签），故统一用 json 标签。
type Config struct {
	HTTPAddr    string            `json:"httpAddr"`
	ChatRpc     zrpc.RpcClientConf `json:"chatRpc"` // 经 etcd 发现 chatsvc
	SigningKeys []string          `json:"signingKeys"` // 司机(driversvc)/乘客(usersvc)JWT 密钥，多密钥兼容
}
