package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDoesNotRequireMySQLDSN 验证网关只依赖 adminsvc RPC，不再要求本地 MySQL DSN。
func TestLoadDoesNotRequireMySQLDSN(t *testing.T) {
	path := writeTestConfig(t, `{}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AdminRPC.Target != "127.0.0.1:8084" {
		t.Fatalf("admin rpc target = %q, want default target", cfg.AdminRPC.Target)
	}
}

// writeTestConfig 写入仅供配置加载测试使用的临时 JSON 文件。
func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "admin.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	return path
}
