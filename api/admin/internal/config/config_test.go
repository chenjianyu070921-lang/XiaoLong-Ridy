package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadPrefersEnvironmentDSN 验证环境变量覆盖配置文件，防止真实凭据重新写回仓库配置。
func TestLoadPrefersEnvironmentDSN(t *testing.T) {
	t.Setenv("ADMIN_API_MYSQL_DSN", "env-dsn")
	path := writeTestConfig(t, `{"mysql":{"dsn":"file-dsn"}}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.MySQL.DSN != "env-dsn" {
		t.Fatalf("dsn = %q, want environment value", cfg.MySQL.DSN)
	}
}

// TestLoadRejectsMissingDSN 验证未配置凭据时服务在启动前失败，而不是用空连接串继续运行。
func TestLoadRejectsMissingDSN(t *testing.T) {
	t.Setenv("ADMIN_API_MYSQL_DSN", "")
	path := writeTestConfig(t, `{"mysql":{"dsn":""}}`)
	if _, err := Load(path); err == nil {
		t.Fatal("Load() should reject empty mysql dsn")
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
