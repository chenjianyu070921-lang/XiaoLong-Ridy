package adminservicelogic

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestValidateRegisterOperator 验证管理员新增操作必须由会话中的同一超级管理员发起。
func TestValidateRegisterOperator(t *testing.T) {
	tests := []struct {
		name              string
		operator          *adminRow
		requestOperatorID int64
		wantCode          codes.Code
	}{
		{
			name:              "超级管理员身份一致",
			operator:          &adminRow{ID: 1001, Role: 1},
			requestOperatorID: 1001,
			wantCode:          codes.OK,
		},
		{
			name:              "非超级管理员被拒绝",
			operator:          &adminRow{ID: 1001, Role: 2},
			requestOperatorID: 1001,
			wantCode:          codes.PermissionDenied,
		},
		{
			name:              "伪造其他操作者ID被拒绝",
			operator:          &adminRow{ID: 1001, Role: 1},
			requestOperatorID: 1002,
			wantCode:          codes.PermissionDenied,
		},
		{
			name:              "缺失会话操作者被拒绝",
			operator:          nil,
			requestOperatorID: 1001,
			wantCode:          codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterOperator(tt.operator, tt.requestOperatorID)
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("validateRegisterOperator() code = %v, want %v", got, tt.wantCode)
			}
		})
	}
}
