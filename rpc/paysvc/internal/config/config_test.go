package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// TestConfigLoadsRefundRedis 验证支付服务业务 Redis 配置不会和 go-zero RPC 内置 Redis 字段冲突。
func TestConfigLoadsRefundRedis(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "paysvc.yaml")
	if err := os.WriteFile(path, []byte(`
Name: paysvc.rpc
ListenOn: 127.0.0.1:0
mysql:
  dsn: root:password@tcp(127.0.0.1:3306)/xiaolong_ridy
refundRedis:
  host: 127.0.0.1:6379
  pass: ""
  db: 1
  poolSize: 10
  dialTimeout: 0
  readTimeout: 0
  writeTimeout: 0
alipay:
  appId: ""
  privateKey: ""
  alipayPublicKey: ""
  gateway: https://openapi.alipay.com/gateway.do
  notifyUrl: ""
  returnUrl: ""
  signType: RSA2
  charset: utf-8
  timeoutExpress: 30m
  sandbox: true
kafka:
  brokers: []
  topic: orderclient.paid
ordersvc:
  target: 127.0.0.1:50051
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var cfg Config
	if err := conf.LoadConfig(path, &cfg); err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RefundRedis.Host != "127.0.0.1:6379" || cfg.RefundRedis.Db != 1 {
		t.Fatalf("RefundRedis = %+v", cfg.RefundRedis)
	}
}
