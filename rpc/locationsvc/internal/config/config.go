package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	MapService MapServiceConfig
}

type MapServiceConfig struct {
	ApiKey   string
	Provider string
	BaseUrl  string
}
