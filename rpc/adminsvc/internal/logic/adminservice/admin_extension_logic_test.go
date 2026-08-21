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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	mock.ExpectExec(`INSERT INTO admin_export_task`).
		WithArgs(sqlmock.AnyArg(), "orders", `{"status":5}`, int64(9001), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(88, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "export", "create", "orders", int64(0), sqlmock.AnyArg(), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))

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
	time.Sleep(20 * time.Millisecond)
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

// TestCreateCoupon_AuditFailureRollsBack 验证优惠券和审计日志使用同一事务。
// 审计写入失败时必须回滚优惠券插入，避免产生无审计的敏感营销配置。
func TestCreateCoupon_AuditFailureRollsBack(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO coupon`).WillReturnResult(sqlmock.NewResult(101, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).WillReturnError(errors.New("audit unavailable"))
	mock.ExpectRollback()

	resp, err := NewCreateCouponLogic(context.Background(), svcCtx).CreateCoupon(&adminsvc.CouponRequest{
		Name: "新用户券", Type: 1, Status: 1, FaceValue: "8.00", Discount: "1.00", ThresholdAmount: "0.00",
		TotalCount: 100, PerUserLimit: 1, ValidStartAt: "2026-08-20 00:00:00", ValidEndAt: "2026-08-21 00:00:00",
		AdminId: 9001, Ip: "127.0.0.1",
	})
	if err == nil {
		t.Fatal("CreateCoupon() error = nil, want audit error")
	}
	if resp != nil {
		t.Fatalf("CreateCoupon() response = %#v, want nil", resp)
	}
}

// TestStatistics_RejectsCityCode 验证订单没有权威城市字段时拒绝城市维度统计，不能静默返回跨城市数据。
func TestStatistics_RejectsCityCode(t *testing.T) {
	for _, call := range []func(*adminsvc.StatisticsRequest) error{
		func(in *adminsvc.StatisticsRequest) error {
			_, err := NewGetStatisticsOverviewLogic(context.Background(), nil).GetStatisticsOverview(in)
			return err
		},
		func(in *adminsvc.StatisticsRequest) error {
			_, err := NewGetOrderStatisticsLogic(context.Background(), nil).GetOrderStatistics(in)
			return err
		},
	} {
		if err := call(&adminsvc.StatisticsRequest{CityCode: "110100"}); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("statistics city_code error code = %v, want %v", status.Code(err), codes.InvalidArgument)
		}
	}
}

// TestParseOrderExportFilters_RejectsUnsafeConditions 验证未知字段和城市字段不会退化为全量订单导出。
func TestParseOrderExportFilters_RejectsUnsafeConditions(t *testing.T) {
	for _, raw := range []string{`{"city_code":"110100"}`, `{"unknown":1}`, `{"status":"5"}`, `{"start_time":"bad"}`} {
		if _, err := parseOrderExportFilters("orders", raw); status.Code(err) != codes.InvalidArgument {
			t.Fatalf("parseOrderExportFilters(%s) code = %v, want %v", raw, status.Code(err), codes.InvalidArgument)
		}
	}
	filters, err := parseOrderExportFilters("orders", `{"status":5,"user_id":1001,"start_time":"2026-08-20 00:00:00","end_time":"2026-08-21 00:00:00"}`)
	if err != nil {
		t.Fatalf("parseOrderExportFilters() error = %v", err)
	}
	where, args := filters.where()
	if !strings.Contains(where, "status = ?") || !strings.Contains(where, "user_id = ?") || len(args) != 4 {
		t.Fatalf("filters.where() = (%q, %#v), want parameterized filter conditions", where, args)
	}
}
