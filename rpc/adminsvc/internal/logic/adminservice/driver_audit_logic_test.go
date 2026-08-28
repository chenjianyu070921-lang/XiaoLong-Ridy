package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeDriversClient struct {
	driverproto.DriverServiceClient

	approveReq *driverproto.AuditCertificationRequest
	rejectReq  *driverproto.AuditCertificationRequest
	listReq    *driverproto.ListDriversRequest
	getReq     *driverproto.GetDriverRequest
	freezeReq  *driverproto.FreezeDriverRequest
	listResp   *driverproto.ListDriversResponse
	getResp    *driverproto.GetDriverResponse
	approveErr error
	rejectErr  error
	listErr    error
	getErr     error
	freezeErr  error
}

func (f *fakeDriversClient) ApproveCertification(ctx context.Context, in *driverproto.AuditCertificationRequest, opts ...grpc.CallOption) (*driverproto.CommonResponse, error) {
	f.approveReq = in
	if f.approveErr != nil {
		return nil, f.approveErr
	}
	return &driverproto.CommonResponse{Message: "ok"}, nil
}

func (f *fakeDriversClient) RejectCertification(ctx context.Context, in *driverproto.AuditCertificationRequest, opts ...grpc.CallOption) (*driverproto.CommonResponse, error) {
	f.rejectReq = in
	if f.rejectErr != nil {
		return nil, f.rejectErr
	}
	return &driverproto.CommonResponse{Message: "ok"}, nil
}

func TestApproveDriverCertification_CallsDriverServiceAndWritesAuditLog(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriverSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "driver", "approve", "driver_certification", int64(3001), "driver certification approved and synced to driver service", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))

	resp, err := NewApproveDriverCertificationLogic(context.Background(), svcCtx).ApproveDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "documents complete",
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
		t.Fatal("ApproveDriverCertification() did not call driver service")
	}
	if driverClient.approveReq.GetCertificationId() != 3001 || driverClient.approveReq.GetOperatorId() != 9001 || driverClient.approveReq.GetRemark() != "documents complete" || driverClient.approveReq.GetIp() != "127.0.0.1" {
		t.Fatalf("driver service approve request = %#v, want certification/operator/remark/ip", driverClient.approveReq)
	}
}

func TestApproveDriverCertification_DownstreamDisabled(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.DriverSvc = nil

	_, err := NewApproveDriverCertificationLogic(context.Background(), svcCtx).ApproveDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		AdminId: 9001,
	})
	if err == nil {
		t.Fatal("expected error when driver service client is nil")
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("want FailedPrecondition, got %v", status.Code(err))
	}
}

func TestRejectDriverCertification_CallsDriverServiceAndWritesAuditLog(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriverSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "driver", "reject", "driver_certification", int64(3001), "driver certification rejected and synced to driver service", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))

	resp, err := NewRejectDriverCertificationLogic(context.Background(), svcCtx).RejectDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "photo is unclear",
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
		t.Fatal("RejectDriverCertification() did not call driver service")
	}
	if driverClient.rejectReq.GetCertificationId() != 3001 || driverClient.rejectReq.GetOperatorId() != 9001 || driverClient.rejectReq.GetRemark() != "photo is unclear" || driverClient.rejectReq.GetIp() != "127.0.0.1" {
		t.Fatalf("driver service reject request = %#v, want certification/operator/remark/ip", driverClient.rejectReq)
	}
}

func TestApproveDriverCertification_CreatesOutboxAndReturnsSuccess(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriverSvc = driverClient

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "driver", "approve", "driver_certification", int64(3001), int64(9001), "driver certification approved and synced to driver service", "127.0.0.1", "operation log write failed").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewApproveDriverCertificationLogic(context.Background(), svcCtx).ApproveDriverCertification(&adminsvc.AuditDriverCertificationRequest{
		Id:      3001,
		Remark:  "documents complete",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil || resp.GetMessage() != "ok" {
		t.Fatalf("ApproveDriverCertification() = %#v, %v; want successful compensated response", resp, err)
	}
	if driverClient.approveReq == nil {
		t.Fatal("ApproveDriverCertification() should call driver service before audit log failure is observed")
	}
}

// TestFreezeDriver_PushDisabledWritesOutboxAndReturnsSuccess 验证冻结主链路成功后，
// pushsvc 未配置不会导致接口失败，而是写入补偿任务等待后续异步通知。
func TestFreezeDriver_PushDisabledWritesOutboxAndReturnsSuccess(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	driverClient := &fakeDriversClient{}
	svcCtx.DriverSvc = driverClient
	svcCtx.PushSvc = nil

	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "driver", "freeze", "driver", int64(2001), "冻结司机：严重违规", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "driver", "freeze_notify", "driver", int64(2001), int64(9001), "严重违规", "127.0.0.1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewFreezeDriverLogic(context.Background(), svcCtx).FreezeDriver(&adminsvc.FreezeDriverRequest{
		Id:      2001,
		Reason:  "严重违规",
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil || resp.GetMessage() != "ok" {
		t.Fatalf("FreezeDriver() = %#v, %v; want success with notification outbox", resp, err)
	}
	if driverClient.freezeReq == nil || driverClient.freezeReq.GetDriverId() != 2001 {
		t.Fatalf("FreezeDriver() downstream request = %#v, want driver 2001", driverClient.freezeReq)
	}
}
