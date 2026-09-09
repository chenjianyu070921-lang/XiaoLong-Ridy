package config

// MysqlConf 通用 mysql 配置结构体，所有服务直接嵌入
type MysqlConf struct {
	Dsn         string `yaml:"dsn" json:"dsn"`
	MaxOpenConn int    `yaml:"maxOpenConn" json:"maxOpenConn,optional"`
	MaxIdleConn int    `yaml:"maxIdleConn" json:"maxIdleConn,optional"`
	MaxLifeTime int    `yaml:"maxLifeTime" json:"maxLifeTime,optional"`
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
