package config

import "time"

// MysqlConf 通用mysql配置结构体，所有服务直接嵌入
type MysqlConf struct {
	Dsn         string        `json:"dsn"`
	MaxOpenConn int           `json:"maxOpenConn"`
	MaxIdleConn int           `json:"maxIdleConn"`
	MaxLifeTime time.Duration `json:"maxLifeTime"`
}

// RedisConf 通用redis配置结构体
type RedisConf struct {
	Host     string `json:"host"`
	Pass     string `json:"pass"`
	Db       int    `json:"db"`
	PoolSize int    `json:"poolSize"`
}
