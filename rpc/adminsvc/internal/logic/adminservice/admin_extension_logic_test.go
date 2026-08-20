package adminservicelogic

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/DATA-DOG/go-sqlmock"
)

// newAdminSQLMock 创建 adminsvc 逻辑层测试使用的 sqlmock 数据库。
// 所有用例都只断言 SQL 行为，不连接真实 MySQL，避免污染业务数据。
func newAdminSQLMock(t *testing.T) (*svc.ServiceContext, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	cleanup := func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
		_ = db.Close()
	}
	return &svc.ServiceContext{MySQL: db}, mock, cleanup
}

// TestCreateExportTask_StartsAsyncStateMachine 验证导出任务会写入 admin_export_task 并启动后台 CSV 生成状态机。
// 任务状态不再写入 admin_operation_log.detail，操作日志只保留审计用途。
func TestCreateExportTask_StartsAsyncStateMachine(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	_ = os.RemoveAll(".tmp-admin-exports")
	defer func() { _ = os.RemoveAll(".tmp-admin-exports") }()

	createdAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local)
	mock.ExpectExec(`INSERT INTO admin_export_task`).
		WithArgs(sqlmock.AnyArg(), "orders", `{"status":5}`, int64(9001), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(88, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "export", "create", "orders", int64(0), sqlmock.AnyArg(), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectQuery(`SELECT task_no, export_type, COALESCE\(CAST\(filters AS CHAR\), ''\), status, admin_id,\s+file_path, file_url, failure_reason, created_at, updated_at, expires_at\s+FROM admin_export_task\s+WHERE task_no = \?`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"task_no", "export_type", "filters", "status", "admin_id", "file_path", "file_url", "failure_reason", "created_at", "updated_at", "expires_at"}).
			AddRow("EX20260820120000000001", "orders", `{"status":5}`, "pending", int64(9001), "", "", "", createdAt, createdAt, nil))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("running", "", "", "", nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT id, order_no, user_id, driver_id, status, estimated_price, created_at\s+FROM ride_order\s+ORDER BY id DESC\s+LIMIT 5000`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_no", "user_id", "driver_id", "status", "estimated_price", "created_at"}).
			AddRow(int64(1001), "RO202608200001", int64(2001), int64(3001), int32(5), "28.50", createdAt))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("success", sqlmock.AnyArg(), sqlmock.AnyArg(), "", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := NewCreateExportTaskLogic(context.Background(), svcCtx).CreateExportTask(&adminsvc.ExportTaskRequest{
		ExportType: "orders",
		Filters:    `{"status":5}`,
		AdminId:    9001,
		Ip:         "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("CreateExportTask() error = %v", err)
	}
	if resp.GetTaskNo() == "" || resp.GetStatus() != "pending" {
		t.Fatalf("CreateExportTask() response = %#v, want pending task", resp)
	}
	time.Sleep(100 * time.Millisecond)
}

// TestGetExportTask_ReturnsFileAndFailureFields 验证导出任务详情会返回文件路径、失败原因和更新时间字段。
func TestGetExportTask_ReturnsFileAndFailureFields(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT task_no, export_type, COALESCE\(CAST\(filters AS CHAR\), ''\), status, admin_id,\s+file_path, file_url, failure_reason, created_at, updated_at, expires_at\s+FROM admin_export_task\s+WHERE task_no = \?`).
		WithArgs("EX20260820120000000001").
		WillReturnRows(sqlmock.NewRows([]string{"task_no", "export_type", "filters", "status", "admin_id", "file_path", "file_url", "failure_reason", "created_at", "updated_at", "expires_at"}).
			AddRow("EX20260820120000000001", "orders", `{"status":5}`, "success", int64(9001), ".tmp-admin-exports/EX20260820120000000001.csv", ".tmp-admin-exports/EX20260820120000000001.csv", "", time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local), time.Date(2026, 8, 20, 12, 0, 1, 0, time.Local), time.Date(2026, 8, 27, 12, 0, 1, 0, time.Local)))

	task, err := NewGetExportTaskLogic(context.Background(), svcCtx).GetExportTask(&adminsvc.ExportTaskDetailRequest{TaskNo: "EX20260820120000000001"})
	if err != nil {
		t.Fatalf("GetExportTask() error = %v", err)
	}
	if task.GetStatus() != "success" || !strings.HasSuffix(task.GetFilePath(), ".csv") || task.GetUpdatedAt() == "" {
		t.Fatalf("GetExportTask() = %#v, want success with file path and updated_at", task)
	}
}

// TestGetStatisticsOverview_ReturnsSQLError 验证统计接口不能吞掉数据库错误。
// 后台看板数据一旦查询失败，应返回错误而不是 200 + 0 值。
func TestGetStatisticsOverview_ReturnsSQLError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM user WHERE deleted_at IS NULL`).
		WillReturnError(errors.New("database unavailable"))

	_, err := NewGetStatisticsOverviewLogic(context.Background(), svcCtx).GetStatisticsOverview(&adminsvc.StatisticsRequest{})
	if err == nil {
		t.Fatal("GetStatisticsOverview() error = nil, want database error")
	}
}

// TestAddBlacklist_ReturnsOperationLogError 验证黑名单新增时审计日志失败必须向上返回。
// 风控黑名单属于敏感操作，不能出现业务成功但 admin_operation_log 缺失的假成功。
func TestAddBlacklist_ReturnsOperationLogError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectExec(`INSERT INTO blacklist`).
		WithArgs("user", int64(1001), "恶意取消订单", int64(9001)).
		WillReturnResult(sqlmock.NewResult(88, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))

	_, err := NewAddBlacklistLogic(context.Background(), svcCtx).AddBlacklist(&adminsvc.BlacklistRequest{
		TargetType: "user",
		TargetId:   1001,
		Reason:     "恶意取消订单",
		AdminId:    9001,
		Ip:         "127.0.0.1",
	})
	if err == nil {
		t.Fatal("AddBlacklist() error = nil, want operation log error")
	}
}

// TestDisableCoupon_UpdatesStatusToStopped 验证优惠券下架使用当前接口文档约定的停用状态 3。
func TestDisableCoupon_UpdatesStatusToStopped(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE coupon\s+SET status = 3, updated_at = \?\s+WHERE id = \? AND status <> 3`).
		WithArgs(sqlmock.AnyArg(), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "coupon", "disable", "coupon", int64(10), "下架优惠券模板：", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectCommit()

	resp, err := NewDisableCouponLogic(context.Background(), svcCtx).DisableCoupon(&adminsvc.CouponRequest{
		Id:      10,
		AdminId: 9001,
		Ip:      "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("DisableCoupon() error = %v", err)
	}
	if resp == nil || resp.Message != "ok" {
		t.Fatalf("DisableCoupon() response = %#v, want ok", resp)
	}
}

// TestValidateCouponRequest_RejectsUnknownStatus 验证优惠券状态只允许当前约定范围内的值。
func TestValidateCouponRequest_RejectsUnknownStatus(t *testing.T) {
	err := validateCouponRequest(&adminsvc.CouponRequest{
		Name:            "测试券",
		Type:            1,
		Status:          4,
		FaceValue:       "8.00",
		Discount:        "1.00",
		ThresholdAmount: "0.00",
		TotalCount:      100,
		PerUserLimit:    1,
		ValidStartAt:    "2026-08-19 00:00:00",
		ValidEndAt:      "2026-08-20 00:00:00",
	})
	if err == nil {
		t.Fatal("validateCouponRequest() error = nil, want invalid status error")
	}
}
