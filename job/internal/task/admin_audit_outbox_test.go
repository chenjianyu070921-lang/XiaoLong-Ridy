package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"XiaoLong-Ridy/job/internal/config"
	"XiaoLong-Ridy/job/internal/svc"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"
	pushproto "XiaoLong-Ridy/rpc/pushesvc/pushesvc"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// fakeDriverClient 提供 outbox 司机冻结补偿测试替身。
type fakeDriverClient struct {
	driverproto.DriverServiceClient
	req *driverproto.FreezeDriverRequest
	err error
}

// FreezeDriver 记录补偿任务重放给 driversvc 的冻结请求。
func (f *fakeDriverClient) FreezeDriver(_ context.Context, req *driverproto.FreezeDriverRequest, _ ...grpc.CallOption) (*driverproto.CommonResponse, error) {
	f.req = req
	if f.err != nil {
		return nil, f.err
	}
	return &driverproto.CommonResponse{Message: "ok"}, nil
}

// fakePushClient 提供 outbox 通知补偿测试替身。
type fakePushClient struct {
	pushproto.PushServiceClient
	noticeReq *pushproto.SendNoticeReq
	pushReq   *pushproto.SendPushReq
}

// SendNotice 记录站内信补偿请求。
func (f *fakePushClient) SendNotice(_ context.Context, req *pushproto.SendNoticeReq, _ ...grpc.CallOption) (*pushproto.SendNoticeResp, error) {
	f.noticeReq = req
	return &pushproto.SendNoticeResp{NoticeId: 1}, nil
}

// SendPush 记录 App 推送补偿请求。
func (f *fakePushClient) SendPush(_ context.Context, req *pushproto.SendPushReq, _ ...grpc.CallOption) (*pushproto.SendPushResp, error) {
	f.pushReq = req
	return &pushproto.SendPushResp{Success: true}, nil
}

// SendSMS 满足 pushsvc 客户端接口，outbox 补偿任务不会调用短信能力。
func (f *fakePushClient) SendSMS(context.Context, *pushproto.SendSMSReq, ...grpc.CallOption) (*pushproto.SendSMSResp, error) {
	return &pushproto.SendSMSResp{Success: true}, nil
}

// newOutboxTaskMock 创建基于 sqlmock 的 GORM 任务测试上下文，避免连接真实数据库。
func newOutboxTaskMock(t *testing.T) (*Task, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
		_ = db.Close()
	}
	return NewTask(&svc.ServiceContext{Config: config.Config{AdminOutboxMaxRetry: 2}, Db: gdb}), mock, cleanup
}

// TestRetryAdminAuditOutbox_ReplaysOperationLog 验证普通审计 outbox 会重写 admin_operation_log 并标记 success。
func TestRetryAdminAuditOutbox_ReplaysOperationLog(t *testing.T) {
	task, mock, cleanup := newOutboxTaskMock(t)
	defer cleanup()
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM `admin_audit_outbox`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_no", "module", "action", "target_type", "target_id", "admin_id", "detail", "ip", "status", "retry_count", "failure_reason", "created_at", "updated_at"}).
			AddRow(1, "AO1", "user", "view_sensitive", "user", 1001, 9001, "查看用户完整手机号和身份证号", "127.0.0.1", "pending", 0, "audit failed", now, now))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `admin_audit_outbox` SET .*`status`=\\?,`updated_at`=\\?").
		WithArgs("running", sqlmock.AnyArg(), int64(1), "pending", "running").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectExec("INSERT INTO admin_operation_log").
		WithArgs(int64(9001), "user", "view_sensitive", "user", int64(1001), "查看用户完整手机号和身份证号", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `admin_audit_outbox` SET .*`failure_reason`=\\?,`status`=\\?,`updated_at`=\\?").
		WithArgs("", "success", sqlmock.AnyArg(), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := task.RetryAdminAuditOutbox(10); err != nil {
		t.Fatalf("RetryAdminAuditOutbox() error = %v", err)
	}
}

// TestRetryAdminAuditOutbox_FreezeFailureMarksFailed 验证冻结补偿失败达到最大次数后会置为 failed。
func TestRetryAdminAuditOutbox_FreezeFailureMarksFailed(t *testing.T) {
	task, mock, cleanup := newOutboxTaskMock(t)
	defer cleanup()
	driver := &fakeDriverClient{err: errors.New("driversvc unavailable")}
	task.svcCtx.DriverClient = driver
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM `admin_audit_outbox`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_no", "module", "action", "target_type", "target_id", "admin_id", "detail", "ip", "status", "retry_count", "failure_reason", "created_at", "updated_at"}).
			AddRow(2, "AO2", "risk", "freeze_driver", "driver", 2001, 9001, "风控黑名单联动冻结司机：高危刷单", "127.0.0.1", "pending", 1, "last failed", now, now))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `admin_audit_outbox` SET .*`status`=\\?,`updated_at`=\\?").
		WithArgs("running", sqlmock.AnyArg(), int64(2), "pending", "running").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `admin_audit_outbox` SET .*`failure_reason`=\\?,`retry_count`=\\?,`status`=\\?,`updated_at`=\\?").
		WithArgs("driversvc unavailable", 2, "failed", sqlmock.AnyArg(), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := task.RetryAdminAuditOutbox(10); err != nil {
		t.Fatalf("RetryAdminAuditOutbox() error = %v", err)
	}
	if driver.req == nil || driver.req.GetDriverId() != 2001 || driver.req.GetOperatorId() != 9001 {
		t.Fatalf("FreezeDriver request = %+v, want driver/operator", driver.req)
	}
}
