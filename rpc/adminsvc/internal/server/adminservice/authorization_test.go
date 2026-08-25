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

// TestRequestAdminIDMatches 验证敏感请求不能使用其他管理员 ID 伪造操作审计身份。
func TestRequestAdminIDMatches(t *testing.T) {
	if !requestAdminIDMatches(&adminsvc.ChangeUserStatusRequest{AdminId: 7}, 7) {
		t.Fatal("matching admin_id should pass")
	}
	if requestAdminIDMatches(&adminsvc.ChangeUserStatusRequest{AdminId: 8}, 7) {
		t.Fatal("forged admin_id should be rejected")
	}
	if !requestAdminIDMatches(&adminsvc.UserListRequest{}, 7) {
		t.Fatal("read request without admin_id should not fail identity matching")
	}
}
