package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 driversvc 的全局配置结构体，映射 etc/driversvc.yaml。
type Config struct {
	zrpc.RpcServerConf // RpcServerConf：go-zero RPC 服务配置（监听地址、Etcd、日志等）

	Mysql     MysqlConf     `yaml:"mysql" json:"mysql"`           // Mysql：MySQL 数据库连接配置
	DriverRedis RedisConf   `yaml:"driverRedis" json:"driverRedis"` // DriverRedis：司机在线状态 / 多端互踢存储（避免与内嵌 RpcServerConf.Redis 冲突）
	Minio     MinioConf     `yaml:"minio" json:"minio"`           // Minio：司机资质图片对象存储
	SigningKey string       `yaml:"signingKey" json:"signingKey"` // SigningKey：JWT 签发/校验的 HMAC-SHA256 密钥
}

// MysqlConf 描述 MySQL 数据源配置。
type MysqlConf struct {
	DSN string `yaml:"dsn" json:"dsn"` // DSN：MySQL 数据源连接串
}

// RedisConf 描述 Redis 连接配置，用于维护司机在线状态与多端互踢。
type RedisConf struct {
	Host     string `yaml:"host" json:"host"`         // Host：Redis 地址（ip:port）
	Password string `yaml:"password" json:"password"` // Password：访问密码，空表示无密码
	DB       int    `yaml:"db" json:"db"`             // DB：逻辑库编号
}

// MinioConf 描述 MinIO 对象存储配置，用于司机资质图片上传。
type MinioConf struct {
	Endpoint  string `yaml:"endpoint" json:"endpoint"`   // Endpoint：MinIO 服务地址（ip:port，不含协议）
	AccessKey string `yaml:"accessKey" json:"accessKey"` // AccessKey：访问账号
	SecretKey string `yaml:"secretKey" json:"secretKey"` // SecretKey：访问密钥
	Bucket    string `yaml:"bucket" json:"bucket"`       // Bucket：资质文件所在桶名
	UseSSL    bool   `yaml:"useSSL" json:"useSSL"`       // UseSSL：是否启用 HTTPS 访问
}
