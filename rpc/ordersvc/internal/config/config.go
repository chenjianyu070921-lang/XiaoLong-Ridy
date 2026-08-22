package config

import (
	cfg "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql       MysqlConf          `yaml:"mysql" json:"mysql"`
	Redis       cfg.RedisConf      `yaml:"myredis" json:"myredis"`
	DispatchRPC zrpc.RpcClientConf `yaml:"dispatchrpc" json:"dispatchrpc"`
	PriceRPC    zrpc.RpcClientConf `yaml:"pricerpc" json:"pricerpc"`
	PayRPC      zrpc.RpcClientConf `yaml:"payrpc" json:"payrpc"`

	// PayChannel 默认支付渠道（1=微信，2=支付宝，3=余额），FinishTrip 生成支付单时使用
	PayChannel int32 `yaml:"paychannel" json:"paychannel"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
