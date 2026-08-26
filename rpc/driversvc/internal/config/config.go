package config

import "github.com/zeromicro/go-zero/zrpc"

// Config maps rpc/driversvc/etc/driversvc.yaml.
type Config struct {
	zrpc.RpcServerConf

	Mysql       MysqlConf       `yaml:"mysql" json:"mysql"`
	SigningKey  string          `yaml:"signingKey" json:"signingKey"`
	DriverRedis DriverRedisConf `yaml:"driverRedis" json:"driverRedis"`
	Minio       MinioConf       `yaml:"minio" json:"minio"`
}

// MysqlConf describes the MySQL datasource.
type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"`
}

// DriverRedisConf describes the Redis instance used by driver online state.
type DriverRedisConf struct {
	Host     string `yaml:"host" json:"host"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

// MinioConf describes the object storage used by driver certification images.
type MinioConf struct {
	Endpoint  string `yaml:"endpoint" json:"endpoint"`
	AccessKey string `yaml:"accessKey" json:"accessKey"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	Bucket    string `yaml:"bucket" json:"bucket"`
	UseSSL    bool   `yaml:"useSSL" json:"useSSL"`
}
