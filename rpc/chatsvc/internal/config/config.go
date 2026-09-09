package config

import "github.com/zeromicro/go-zero/zrpc"

// Config maps rpc/chatsvc/etc/chatsvc.yaml.
type Config struct {
	zrpc.RpcServerConf

	Mysql         MysqlConf         `yaml:"mysql" json:"mysql"`
	DriverRedis   DriverRedisConf   `yaml:"driverRedis" json:"driverRedis"`
	OrderRpc      zrpc.RpcClientConf `yaml:"orderRpc" json:"orderRpc"`
	SensitiveWords []string         `yaml:"sensitiveWords" json:"sensitiveWords"`
}

// MysqlConf describes the MySQL datasource.
type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

// DriverRedisConf describes the Redis instance used for driver push (driver:push:%d).
type DriverRedisConf struct {
	Host     string `yaml:"host" json:"host"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}
