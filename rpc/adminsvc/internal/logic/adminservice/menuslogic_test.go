package adminservicelogic

import (
	"testing"

	adminconfig "XiaoLong-Ridy/rpc/adminsvc/internal/config"
)

// TestMapMenus_ReadsRoleConfiguration 验证菜单由角色配置驱动，而非在逻辑层固定角色分支。
func TestMapMenus_ReadsRoleConfiguration(t *testing.T) {
	menuRoles := map[int32][]adminconfig.MenuItemConfig{
		1: {{Name: "管理员", Path: "/admins", Perm: "admin:manage"}},
		2: {{Name: "订单监控", Path: "/orders", Perm: "order:list"}},
	}

	items := mapMenus(2, menuRoles)
	if len(items) != 1 || items[0].GetPath() != "/orders" || items[0].GetPerm() != "order:list" {
		t.Fatalf("mapMenus() = %#v, want configured role 2 menu", items)
	}
	if items := mapMenus(3, menuRoles); len(items) != 0 {
		t.Fatalf("mapMenus() = %#v, want no menu for unconfigured role", items)
	}
}
