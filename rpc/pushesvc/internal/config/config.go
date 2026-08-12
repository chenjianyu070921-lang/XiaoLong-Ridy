package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	SMS  SMSConfig
	Push PushConfig
}

type SMSConfig struct {
	Provider  string
	AccessKey string
	SecretKey string
	SignName  string
}

type PushConfig struct {
	Provider     string
	AppKey       string
	MasterSecret string
}
