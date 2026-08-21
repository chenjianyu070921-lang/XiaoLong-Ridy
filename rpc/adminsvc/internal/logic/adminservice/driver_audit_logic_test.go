package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	driversvcproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
)

// fakeDriversClient 是 adminsvc 司机审核联调用例中的 driversvc 假客户端。
// 测试只关心审核 RPC 参数是否正确透传，其余 driversvc 方法通过嵌入接口占位。
type fakeDriversClient struct {
	driversvcproto.DriversvcClient

	approveReq *driversvcproto.AuditCertificationRequest
	rejectReq  *driversvcproto.AuditCertificationRequest
	approveErr error
	rejectErr  error
}

// ApproveCertification 记录审核通过请求，模拟 driversvc 成功或失败。
func (f *fakeDriversClient) ApproveCertification(ctx context.Context, in *driversvcproto.AuditCertificationRequest, opts ...grpc.CallOption) (*driversvcproto.CommonResponse, error) {
	f.approveReq = in
	if f.approveErr != nil {
		return nil, f.approveErr
	}
	return &driversvcproto.CommonResponse{Message: "ok"}, nil
}

// RejectCertification 记录审核驳回请求，模拟 driversvc 成功或失败。
func (f *fakeDriversClient) RejectCertification(ctx context.Context, in *driversvcproto.AuditCertificationRequest, opts ...grpc.CallOption) (*driversvcproto.CommonResponse, error) {
	f.rejectReq = in
	if f.rejectErr != nil {
		return nil, f.rejectErr
	}
	return &driversvcproto.CommonResponse{Message: "ok"}, nil
}

// TestApproveDriverCertification_CallsDriversvcAndWritesAuditLog 验证审核通过会调用 driversvc 并写入审计日志。
func TestApproveDriverCertification_CallsDriversvcAndWritesAuditLog(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriversSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "driver", "approve", "driver_certification", int64(3001), "司机认证通过，已同步 driversvc 并联动司机可听单状态", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))

	resp, err := NewApproveDriverCertificationLogic(context.Background(), svcCtx).ApproveDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "资料齐全",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("ApproveDriverCertification() error = %v", err)
	}
	if resp == nil || resp.GetMessage() != "ok" {
		t.Fatalf("ApproveDriverCertification() response = %#v, want ok", resp)
	}
	if driverClient.approveReq == nil {
		t.Fatal("ApproveDriverCertification() did not call driversvc")
	}
	if driverClient.approveReq.GetCertificationId() != 3001 || driverClient.approveReq.GetOperatorId() != 9001 || driverClient.approveReq.GetRemark() != "资料齐全" || driverClient.approveReq.GetIp() != "127.0.0.1" {
		t.Fatalf("driversvc approve request = %#v, want certification/operator/remark/ip", driverClient.approveReq)
	}
}

// TestRejectDriverCertification_CallsDriversvcAndWritesAuditLog 验证审核驳回会调用 driversvc 并写入审计日志。
func TestRejectDriverCertification_CallsDriversvcAndWritesAuditLog(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriversSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "driver", "reject", "driver_certification", int64(3001), "司机认证驳回，已同步 driversvc", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))

	resp, err := NewRejectDriverCertificationLogic(context.Background(), svcCtx).RejectDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "证件照片模糊",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("RejectDriverCertification() error = %v", err)
	}
	if resp == nil || resp.GetMessage() != "ok" {
		t.Fatalf("RejectDriverCertification() response = %#v, want ok", resp)
	}
	if driverClient.rejectReq == nil {
		t.Fatal("RejectDriverCertification() did not call driversvc")
	}
	if driverClient.rejectReq.GetCertificationId() != 3001 || driverClient.rejectReq.GetOperatorId() != 9001 || driverClient.rejectReq.GetRemark() != "证件照片模糊" || driverClient.rejectReq.GetIp() != "127.0.0.1" {
		t.Fatalf("driversvc reject request = %#v, want certification/operator/remark/ip", driverClient.rejectReq)
	}
}

// TestApproveDriverCertification_ReturnsErrorWhenAuditLogFails 验证敏感审核操作在审计日志失败时不能返回成功。
// driversvc 已完成的跨服务状态变更无法由 adminsvc 本地事务回滚，因此策略是阻断成功响应，并写入 admin_audit_outbox 等待后续补偿审计。
func TestApproveDriverCertification_ReturnsErrorWhenAuditLogFails(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriversSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "driver", "approve", "driver_certification", int64(3001), int64(9001), "司机认证通过，已同步 driversvc 并联动司机可听单状态", "127.0.0.1", "operation log write failed").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewApproveDriverCertificationLogic(context.Background(), svcCtx).ApproveDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "资料齐全",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err == nil {
		t.Fatal("ApproveDriverCertification() error = nil, want audit log error")
	}
	if resp != nil {
		t.Fatalf("ApproveDriverCertification() response = %#v, want nil", resp)
	}
	if driverClient.approveReq == nil {
		t.Fatal("ApproveDriverCertification() should call driversvc before audit log failure is observed")
	}
}
