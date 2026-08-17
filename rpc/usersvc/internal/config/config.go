package config

import "github.com/zeromicro/go-zero/zrpc"

// Config 是 usersvc 的 RPC 服务配置，遵循 goctl 生成的配置承载方式。
type Config struct {
	zrpc.RpcServerConf

	Mysql      MysqlConf `yaml:"mysql" json:"mysql"`
	CacheRedis RedisConf `yaml:"cacheRedis" json:"cacheRedis"`
	SMS        SMSConf   `yaml:"sms" json:"sms"`
	TokenAuth  AuthConf  `yaml:"tokenAuth" json:"tokenAuth"`
}

// MysqlConf 保存 usersvc 持久化用户和地址数据所需的 MySQL 配置。
type MysqlConf struct {
	DSN         string `yaml:"dsn" json:"dsn"`
	MaxOpenConn int    `yaml:"maxOpenConn" json:"maxOpenConn"`
	MaxIdleConn int    `yaml:"maxIdleConn" json:"maxIdleConn"`
	MaxLifeTime int    `yaml:"maxLifeTime" json:"maxLifeTime"`
}

// RedisConf 保存 usersvc 验证码和令牌会话所需的 Redis 配置。
type RedisConf struct {
	Host     string `yaml:"host" json:"host"`
	Pass     string `yaml:"pass" json:"pass"`
	DB       int    `yaml:"db" json:"db"`
	PoolSize int    `yaml:"poolSize" json:"poolSize"`
}

// SMSConf 保存腾讯云短信发送所需的配置。
type SMSConf struct {
	Provider    string `yaml:"provider" json:"provider"`
	Region      string `yaml:"region" json:"region"`
	SecretID    string `yaml:"secretId" json:"secretId"`
	SecretKey   string `yaml:"secretKey" json:"secretKey"`
	SmsSdkAppID string `yaml:"smsSdkAppId" json:"smsSdkAppId"`
	SignName    string `yaml:"signName" json:"signName"`
	TemplateID  string `yaml:"templateId" json:"templateId"`
}

// AuthConf 保存本地开发期令牌签名配置，后续可替换为统一配置中心或密钥服务。
type AuthConf struct {
	SigningKey string `yaml:"signingKey" json:"signingKey"`
}
