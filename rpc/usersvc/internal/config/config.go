package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 usersvc 的 RPC 服务配置，遵循 goctl 生成的配置承载方式。
type Config struct {
	zrpc.RpcServerConf

	TokenAuth AuthConf `yaml:"tokenAuth" json:"tokenAuth"`
}

// AuthConf 保存本地开发期令牌签名配置，后续可替换为统一配置中心或密钥服务。
type AuthConf struct {
	SigningKey string `yaml:"signingKey" json:"signingKey"`
}
