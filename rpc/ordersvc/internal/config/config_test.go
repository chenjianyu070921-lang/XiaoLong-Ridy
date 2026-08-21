package config

import (
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestOrdersvcConfig_LoadsRedisConf 验证 ordersvc 独立 redisconf 可与 go-zero RpcServerConf 同时加载。
// 该测试防止配置字段冲突再次导致服务在创建取消订单 RPC 客户端前启动失败。
func TestOrdersvcConfig_LoadsRedisConf(t *testing.T) {
	configPath := filepath.Join("..", "..", "etc", "ordersvc.yaml")
	var c Config
	if err := conf.Load(configPath, &c); err != nil {
		t.Fatalf("conf.Load(%q): %v", configPath, err)
	}
	if c.ListenOn == "" || c.Mysql.DSN == "" || c.Redis.Host == "" {
		t.Fatalf("ordersvc config missing required values: listen=%q mysql=%t redis=%q", c.ListenOn, c.Mysql.DSN != "", c.Redis.Host)
	}
}
