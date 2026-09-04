package config

// MysqlConf 通用 mysql 配置结构体，所有服务直接嵌入
type MysqlConf struct {
	Dsn         string `json:"dsn"`
	MaxOpenConn int    `json:"maxOpenConn"`
	MaxIdleConn int    `json:"maxIdleConn"`
	MaxLifeTime int    `json:"maxLifeTime"`
}

// RedisConf 通用 redis 配置结构体，所有服务直接嵌入
type RedisConf struct {
	Host         string `json:"host"`
	Pass         string `json:"pass"`
	Db           int    `json:"db"`
	PoolSize     int    `json:"poolSize"`
	DialTimeout  int    `json:"dialTimeout"`  // 秒，<=0 时默认 5
	ReadTimeout  int    `json:"readTimeout"`  // 秒，<=0 时默认 3
	WriteTimeout int    `json:"writeTimeout"` // 秒，<=0 时默认 3
}
