package server

import (
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
)

// TestRoleAllowed 验证后台角色矩阵在 RPC 服务端生效，不能只依赖前端菜单隐藏按钮。
func TestRoleAllowed(t *testing.T) {
	cases := []struct {
		name   string
		method string
		role   int32
		want   bool
	}{
		{name: "超管可以改价", method: "/adminsvc.AdminService/CreatePriceRule", role: adminRoleSuper, want: true},
		{name: "运营不能改价", method: "/adminsvc.AdminService/CreatePriceRule", role: adminRoleOps, want: false},
		{name: "运营可以编辑优惠券", method: "/adminsvc.AdminService/UpdateCoupon", role: adminRoleOps, want: true},
		{name: "运营不能发券", method: "/adminsvc.AdminService/IssueCoupon", role: adminRoleOps, want: false},
		{name: "客服可以取消订单", method: "/adminsvc.AdminService/CancelOrder", role: adminRoleCS, want: true},
		{name: "客服不能冻结用户", method: "/adminsvc.AdminService/FreezeUser", role: adminRoleCS, want: false},
		{name: "运营可以进入工单动作入口", method: "/adminsvc.AdminService/ActWorkOrder", role: adminRoleOps, want: true},
		{name: "客服可以进入工单动作入口", method: "/adminsvc.AdminService/ActWorkOrder", role: adminRoleCS, want: true},
		{name: "运营可以批量处理工单", method: "/adminsvc.AdminService/BatchActWorkOrders", role: adminRoleOps, want: true},
		{name: "客服可以批量跟进工单", method: "/adminsvc.AdminService/BatchActWorkOrders", role: adminRoleCS, want: true},
		{name: "运营可以查询风控命中", method: "/adminsvc.AdminService/ListRiskHitRecords", role: adminRoleOps, want: true},
		{name: "客服不能查询风控命中", method: "/adminsvc.AdminService/ListRiskHitRecords", role: adminRoleCS, want: false},
		{name: "运营不能处置风控命中", method: "/adminsvc.AdminService/HandleRiskHitRecords", role: adminRoleOps, want: false},
		{name: "客服不能处置风控命中", method: "/adminsvc.AdminService/HandleRiskHitRecords", role: adminRoleCS, want: false},
		{name: "未知角色拒绝", method: "/adminsvc.AdminService/ListUsers", role: 99, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := roleAllowed(tc.method, tc.role); got != tc.want {
				t.Fatalf("roleAllowed(%q, %d) = %v, want %v", tc.method, tc.role, got, tc.want)
			}
		})
	}
}

// TestRequestAdminIDMatches 验证操作者身份类写请求不能伪造管理员 ID，
// 同时确认查询目标/过滤类 admin_id 不参与身份一致性校验。
func TestRequestAdminIDMatches(t *testing.T) {
	writeMethod := "/adminsvc.AdminService/FreezeUser"
	readMethods := []string{
		"/adminsvc.AdminService/Me",
		"/adminsvc.AdminService/Menus",
		"/adminsvc.AdminService/ListOperationLogs",
	}

	// 写请求：admin_id 与会话一致时放行。
	if !requestAdminIDMatches(writeMethod, &adminsvc.ChangeUserStatusRequest{AdminId: 7}, 7) {
		t.Fatal("matching admin_id should pass")
	}
	// 写请求：admin_id 与会话不一致时拒绝，防止伪造操作者身份。
	if requestAdminIDMatches(writeMethod, &adminsvc.ChangeUserStatusRequest{AdminId: 8}, 7) {
		t.Fatal("forged admin_id should be rejected")
	}
	// 读请求：即使请求体带有 admin_id 字段（查询目标或过滤条件），也不做身份一致性校验。
	for _, method := range readMethods {
		if !requestAdminIDMatches(method, &adminsvc.MeRequest{AdminId: 0}, 7) {
			t.Fatalf("read request %s with admin_id filter should pass identity matching", method)
		}
	}
	// 无 admin_id 字段的请求始终通过身份一致性校验。
	if !requestAdminIDMatches("/adminsvc.AdminService/ListUsers", &adminsvc.UserListRequest{}, 7) {
		t.Fatal("read request without admin_id should not fail identity matching")
	}
}
