// Package config 负责加载和校验管理后台运行配置。
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/zrpc"
)

// Config 定义管理后台 HTTP 服务、Redis、鉴权和 adminsvc RPC 客户端配置。
type Config struct {
	HTTPAddr string `json:"http_addr"`
	// Redis 为兼容保留字段；当前 api/admin 不直接连接 Redis，
	// 会话与登录凭据统一由 adminsvc 通过环境变量管理，此处不保存任何密码。
	Redis    RedisConfig        `json:"redis"`
	Auth     AuthConfig         `json:"auth"`
	AdminRPC zrpc.RpcClientConf `json:"admin_rpc"`
}

// RedisConfig 定义 Redis 服务地址、密码和逻辑数据库编号。
type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// AuthConfig 定义后台登录会话的过期时间和 Redis 键前缀。
type AuthConfig struct {
	SessionTTLHours int    `json:"session_ttl_hours"`
	TokenPrefix     string `json:"token_prefix"`
}

// Load 从 JSON 文件加载配置，并补齐不会影响本地开发的默认值。
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8717"
	}
	if cfg.Auth.SessionTTLHours <= 0 {
		cfg.Auth.SessionTTLHours = 24
	}
	if cfg.Auth.TokenPrefix == "" {
		cfg.Auth.TokenPrefix = "admin:sess:"
	}
	if len(cfg.AdminRPC.Endpoints) == 0 && cfg.AdminRPC.Target == "" {
		cfg.AdminRPC.Target = "127.0.0.1:8084"
	}
	return &cfg, nil
}
