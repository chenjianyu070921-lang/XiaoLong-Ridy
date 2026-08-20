package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 driversvc 的全局配置结构体，映射 etc/driversvc.yaml。
type Config struct {
	zrpc.RpcServerConf // RpcServerConf：go-zero RPC 服务配置（监听地址、Etcd、日志等）

	Mysql MysqlConf `yaml:"mysql" json:"mysql"` // Mysql：MySQL 数据库连接配置
	SigningKey string `yaml:"signingKey" json:"signingKey"` // SigningKey：JWT 签发/校验的 HMAC-SHA256 密钥
}

// MysqlConf 描述 MySQL 数据源配置。
type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"` // DSN：MySQL 数据源连接串
}
