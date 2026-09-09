package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestYAMLMenuRoles 验证 yaml.v3 能按既有 admin.yaml 的字段名解析 MenuRoles。
// 防止配置字段名与 YAML key 大小写不匹配导致菜单配置被静默丢弃。
func TestYAMLMenuRoles(t *testing.T) {
	sample := `
MySQL:
  DSN: "root:pwd@tcp(127.0.0.1:3306)/db"
Cache:
  Host: 127.0.0.1:6379
  Password: ""
  DB: 0
Session:
  SessionTTLHours: 24
  TokenPrefix: "admin:sess:"
MenuRoles:
  1:
  - Name: 管理员
    Path: /admins
    Icon: Shield
    Perm: admin:manage
`
	var c Config
	if err := yaml.Unmarshal([]byte(sample), &c); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if c.MySQL.DSN == "" {
		t.Error("MySQL.DSN should be parsed from yaml")
	}
	if c.Cache.Host != "127.0.0.1:6379" {
		t.Errorf("Cache.Host = %q, want 127.0.0.1:6379", c.Cache.Host)
	}
	items, ok := c.MenuRoles[1]
	if !ok || len(items) != 1 {
		t.Fatalf("MenuRoles[1] = %+v, want 1 item", c.MenuRoles)
	}
	if items[0].Name != "管理员" || items[0].Path != "/admins" {
		t.Fatalf("unexpected menu item: %+v", items[0])
	}
}
