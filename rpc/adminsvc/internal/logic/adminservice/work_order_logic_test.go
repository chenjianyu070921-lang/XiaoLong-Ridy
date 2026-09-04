package adminservicelogic

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestValidateWorkOrderActionRole 验证工单动作按管理员角色进行最小权限控制。
// 该测试覆盖所有设计中定义的角色和状态动作，确保仲裁、结案、重开不会被普通角色执行。
func TestValidateWorkOrderActionRole(t *testing.T) {
	tests := []struct {
		name   string
		role   int32
		action string
		code   codes.Code
	}{
		{name: "超管可以分配", role: 1, action: "assign", code: codes.OK},
		{name: "超管可以跟进", role: 1, action: "follow", code: codes.OK},
		{name: "超管可以仲裁", role: 1, action: "arbitrate", code: codes.OK},
		{name: "超管可以结案", role: 1, action: "close", code: codes.OK},
		{name: "超管可以重开", role: 1, action: "reopen", code: codes.OK},
		{name: "运营可以分配", role: 2, action: "assign", code: codes.OK},
		{name: "运营可以跟进", role: 2, action: "follow", code: codes.OK},
		{name: "运营不能仲裁", role: 2, action: "arbitrate", code: codes.PermissionDenied},
		{name: "运营不能结案", role: 2, action: "close", code: codes.PermissionDenied},
		{name: "运营不能重开", role: 2, action: "reopen", code: codes.PermissionDenied},
		{name: "客服不能分配", role: 3, action: "assign", code: codes.PermissionDenied},
		{name: "客服可以跟进", role: 3, action: "follow", code: codes.OK},
		{name: "客服不能仲裁", role: 3, action: "arbitrate", code: codes.PermissionDenied},
		{name: "客服不能结案", role: 3, action: "close", code: codes.PermissionDenied},
		{name: "客服不能重开", role: 3, action: "reopen", code: codes.PermissionDenied},
		{name: "未知角色拒绝", role: 99, action: "follow", code: codes.PermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkOrderActionRole(tt.role, tt.action)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("validateWorkOrderActionRole(%d, %q) code = %v, want %v", tt.role, tt.action, got, tt.code)
			}
		})
	}
}
