package config

import (
	"XiaoLong-Ridy/common/alipay"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql  MysqlConf     `yaml:"mysql" json:"mysql"`
	Alipay alipay.Config `yaml:"alipay" json:"alipay"`
}

type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}
