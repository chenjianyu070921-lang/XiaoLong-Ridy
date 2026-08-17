package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestApproveCertificationLogic_ReturnsPreconditionWhenDBMissing 验证数据库未初始化时返回受控错误。
func TestApproveCertificationLogic_ReturnsPreconditionWhenDBMissing(t *testing.T) {
	logic := NewApproveCertificationLogic(context.Background(), &svc.ServiceContext{})
	resp, err := logic.ApproveCertification(&proto.AuditCertificationRequest{
		CertificationId: 1001,
		OperatorId:      9001,
		Remark:          "资料齐全",
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("ApproveCertification() error code = %v, want %v", status.Code(err), codes.FailedPrecondition)
	}
	if resp != nil {
		t.Fatalf("ApproveCertification() response = %#v, want nil", resp)
	}
}
