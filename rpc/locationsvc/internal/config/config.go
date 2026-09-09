package config

import (
	commonconfig "XiaoLong-Ridy/common/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Mysql      commonconfig.MysqlConf
	RedisConf  commonconfig.RedisConf `yaml:"myredis" json:"myredis"`
	MapService MapServiceConfig
	// DefaultCityCode 司机位置 GEO 默认城市编码，需与 dispatchsvc 保持一致。
	DefaultCityCode string `yaml:"defaultCityCode" json:"defaultCityCode"`
}

type MapServiceConfig struct {
	ApiKey   string
	Provider string
	BaseUrl  string
}
