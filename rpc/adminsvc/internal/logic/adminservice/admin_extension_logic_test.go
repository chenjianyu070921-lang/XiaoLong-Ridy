package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
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
	mock.ExpectQuery(`SELECT id, order_no, user_id, driver_id, status, estimated_price, created_at\s+FROM ride_order\s+WHERE 1=1 AND status = \?\s+ORDER BY id DESC\s+LIMIT 5000`).
		WithArgs(int32(5)).
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

// TestRunExportTaskJob_MarksFailedWhenLoadFails 验证任务读取失败时会回写 failed，避免任务永久停留在 pending。
func TestRunExportTaskJob_MarksFailedWhenLoadFails(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT task_no, export_type, COALESCE\(CAST\(filters AS CHAR\), ''\), status, admin_id,\s+file_path, file_url, failure_reason, created_at, updated_at, expires_at\s+FROM admin_export_task\s+WHERE task_no = \?`).
		WithArgs("EXLOADFAILED").
		WillReturnError(errors.New("database unavailable"))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("failed", "", "", sqlmock.AnyArg(), nil, "EXLOADFAILED").
		WillReturnResult(sqlmock.NewResult(0, 1))

	runExportTaskJob(svcCtx, "EXLOADFAILED")
}

// TestRunExportTaskJob_MarksFailedWhenRunningUpdateFails 验证更新 running 失败后会立即尝试写入 failed。
func TestRunExportTaskJob_MarksFailedWhenRunningUpdateFails(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	createdAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local)
	mock.ExpectQuery(`SELECT task_no, export_type, COALESCE\(CAST\(filters AS CHAR\), ''\), status, admin_id,\s+file_path, file_url, failure_reason, created_at, updated_at, expires_at\s+FROM admin_export_task\s+WHERE task_no = \?`).
		WithArgs("EXRUNFAILED").
		WillReturnRows(sqlmock.NewRows([]string{"task_no", "export_type", "filters", "status", "admin_id", "file_path", "file_url", "failure_reason", "created_at", "updated_at", "expires_at"}).
			AddRow("EXRUNFAILED", "orders", `{"status":5}`, "pending", int64(9001), "", "", "", createdAt, createdAt, nil))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("running", "", "", "", nil, "EXRUNFAILED").
		WillReturnError(errors.New("database unavailable"))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("failed", "", "", sqlmock.AnyArg(), nil, "EXRUNFAILED").
		WillReturnResult(sqlmock.NewResult(0, 1))

	runExportTaskJob(svcCtx, "EXRUNFAILED")
}

// TestRunExportTaskJob_RecoversPanicAndMarksFailed 验证 CSV 生成 panic 被恢复，并将任务标记为 failed。
func TestRunExportTaskJob_RecoversPanicAndMarksFailed(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	createdAt := time.Date(2026, 8, 20, 12, 0, 0, 0, time.Local)
	mock.ExpectQuery(`SELECT task_no, export_type, COALESCE\(CAST\(filters AS CHAR\), ''\), status, admin_id,\s+file_path, file_url, failure_reason, created_at, updated_at, expires_at\s+FROM admin_export_task\s+WHERE task_no = \?`).
		WithArgs("EXPANIC").
		WillReturnRows(sqlmock.NewRows([]string{"task_no", "export_type", "filters", "status", "admin_id", "file_path", "file_url", "failure_reason", "created_at", "updated_at", "expires_at"}).
			AddRow("EXPANIC", "orders", `{"status":5}`, "pending", int64(9001), "", "", "", createdAt, createdAt, nil))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("running", "", "", "", nil, "EXPANIC").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE admin_export_task\s+SET status = \?, file_path = \?, file_url = \?, failure_reason = \?, expires_at = \?\s+WHERE task_no = \?`).
		WithArgs("failed", "", "", "导出任务异常: test panic", nil, "EXPANIC").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 替换文件生成器以模拟真实 worker 内部 panic，并使用同一锁避免与异步 worker 竞争。
	exportFileWriterMu.Lock()
	previousWriter := exportFileWriter
	exportFileWriter = func(context.Context, *svc.ServiceContext, *adminsvc.ExportTask) (string, error) {
		panic("test panic")
	}
	exportFileWriterMu.Unlock()
	defer func() {
		exportFileWriterMu.Lock()
		exportFileWriter = previousWriter
		exportFileWriterMu.Unlock()
	}()

	runExportTaskJob(svcCtx, "EXPANIC")
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

// TestAddBlacklist_CreatesOutboxAndReturnsSuccess 验证业务写入成功后审计失败会创建补偿任务并返回成功。
func TestAddBlacklist_CreatesOutboxAndReturnsSuccess(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM blacklist WHERE target_type = \? AND target_id = \? AND status = 1`).
		WithArgs("user", int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO blacklist`).
		WithArgs("user", int64(1001), "恶意取消订单", int64(9001)).
		WillReturnResult(sqlmock.NewResult(88, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WillReturnError(errors.New("operation log write failed"))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "risk", "add_blacklist", "blacklist", int64(88), int64(9001), "新增黑名单：user/1001", "127.0.0.1", "operation log write failed").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewAddBlacklistLogic(context.Background(), svcCtx).AddBlacklist(&adminsvc.BlacklistRequest{
		TargetType: "user",
		TargetId:   1001,
		Reason:     "恶意取消订单",
		AdminId:    9001,
		Ip:         "127.0.0.1",
	})
	if err != nil || resp.GetMessage() != "ok" {
		t.Fatalf("AddBlacklist() = %#v, %v; want successful compensated response", resp, err)
	}
}

// TestAddBlacklist_DriverFreezeFailureWritesOutboxAndReturnsSuccess 验证司机拉黑已提交后，
// driversvc 冻结失败会写入补偿任务，接口仍保持黑名单主处置成功语义。
func TestAddBlacklist_DriverFreezeFailureWritesOutboxAndReturnsSuccess(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.DriverSvc = &fakeDriversClient{freezeErr: errors.New("driversvc unavailable")}

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM blacklist WHERE target_type = \? AND target_id = \? AND status = 1`).
		WithArgs("driver", int64(2001)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO blacklist`).
		WithArgs("driver", int64(2001), "高危刷单", int64(9001)).
		WillReturnResult(sqlmock.NewResult(89, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "risk", "add_blacklist", "blacklist", int64(89), "新增黑名单：driver/2001", "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(99, 1))
	mock.ExpectExec(`INSERT INTO admin_audit_outbox`).
		WithArgs(sqlmock.AnyArg(), "risk", "freeze_driver", "driver", int64(2001), int64(9001), "风控黑名单联动冻结司机：高危刷单", "127.0.0.1", "driversvc unavailable").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := NewAddBlacklistLogic(context.Background(), svcCtx).AddBlacklist(&adminsvc.BlacklistRequest{
		TargetType: "driver",
		TargetId:   2001,
		Reason:     "高危刷单",
		AdminId:    9001,
		Ip:         "127.0.0.1",
	})
	if err != nil || resp.GetMessage() != "ok" {
		t.Fatalf("AddBlacklist() = %#v, %v; want blacklist success with freeze outbox", resp, err)
	}
}

// TestAddBlacklist_RejectsDuplicateActiveTarget 验证新增黑名单前会拒绝重复生效记录，避免同一目标被多次拉黑。
func TestAddBlacklist_RejectsDuplicateActiveTarget(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM blacklist WHERE target_type = \? AND target_id = \? AND status = 1`).
		WithArgs("user", int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := NewAddBlacklistLogic(context.Background(), svcCtx).AddBlacklist(&adminsvc.BlacklistRequest{
		TargetType: "user",
		TargetId:   1001,
		Reason:     "重复拉黑",
		AdminId:    9001,
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("AddBlacklist() error code = %v, want AlreadyExists", status.Code(err))
	}
}

// TestListRiskHitRecords_ReturnsDerivedDisposition 验证风控命中列表会返回由审计、黑名单和工单推导出的处置状态。
// 当前数据库表没有 handle_status 字段，因此这里显式断言服务层聚合字段，避免前端只能看到待处理原始命中。
func TestListRiskHitRecords_ReturnsDerivedDisposition(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	createdAt := time.Date(2026, 8, 28, 10, 0, 0, 0, time.Local)
	handledAt := time.Date(2026, 8, 28, 10, 5, 0, 0, time.Local)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM risk_blacklist_hit_record r WHERE 1=1 AND r.target_type = \?`).
		WithArgs("driver").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT r\.id, r\.blacklist_id, r\.target_type, r\.target_id, r\.scene, r\.risk_level, r\.hit_reason, r\.request_id, r\.created_at,`).
		WithArgs("driver", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "blacklist_id", "target_type", "target_id", "scene", "risk_level", "hit_reason", "request_id", "created_at",
			"work_order_id", "active_blacklist_id", "review_action", "review_admin_id", "review_at",
		}).AddRow(int64(11), int64(0), "driver", int64(2001), "dispatch", int32(3), "高危派单", "req-1", createdAt, int64(900), int64(88), "review_pass", int64(7001), handledAt))

	resp, err := NewListRiskHitRecordsLogic(context.Background(), svcCtx).ListRiskHitRecords(&adminsvc.RiskHitRecordListRequest{
		Page: 1, PageSize: 20, TargetType: "driver",
	})
	if err != nil {
		t.Fatalf("ListRiskHitRecords() error = %v", err)
	}
	if len(resp.GetList()) != 1 {
		t.Fatalf("ListRiskHitRecords() list len = %d, want 1", len(resp.GetList()))
	}
	item := resp.GetList()[0]
	if item.GetHandleStatus() != "work_order" || item.GetHandleAction() != "hit_create_work_order" || item.GetWorkOrderId() != 900 {
		t.Fatalf("RiskHitRecord disposition = %#v, want derived work_order status", item)
	}
}

// TestHandleRiskHitRecords_RejectsOpsAddBlacklist 验证运营角色不能执行高危拉黑动作。
// 风控权限分层保持在 adminsvc 服务端校验，避免客户端绕过 HTTP 菜单或按钮限制。
func TestHandleRiskHitRecords_RejectsOpsAddBlacklist(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	server := miniredis.RunT(t)
	defer server.Close()
	svcCtx.Redis = redis.NewClient(&redis.Options{Addr: server.Addr()})
	mock.ExpectQuery(`SELECT id, username, password_hash, real_name, role, status\s+FROM admin_user\s+WHERE id = \? AND deleted_at IS NULL`).
		WithArgs(int64(9002)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "real_name", "role", "status"}).
			AddRow(int64(9002), "ops", "hash", "运营", int32(2), int32(1)))

	_, err := NewHandleRiskHitRecordsLogic(contextWithAdminSession(t, svcCtx, 9002, 2), svcCtx).HandleRiskHitRecords(&adminsvc.RiskHitActionRequest{
		Ids: []int64{11}, Action: "add_blacklist", Reason: "高危命中", AdminId: 9002,
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("HandleRiskHitRecords() code = %v, want PermissionDenied", status.Code(err))
	}
}

// TestHandleRiskHitRecords_RejectsDuplicateReviewPass 验证重复复核不会再次写审计。
// 批量处置采用单条失败汇总语义，因此重复处置会体现在 fail_count 和 failure_reasons 中。
func TestHandleRiskHitRecords_RejectsDuplicateReviewPass(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	createdAt := time.Date(2026, 8, 28, 10, 0, 0, 0, time.Local)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, blacklist_id, target_type, target_id, scene, risk_level, hit_reason, request_id, created_at\s+FROM risk_blacklist_hit_record\s+WHERE id = \?\s+FOR UPDATE`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "blacklist_id", "target_type", "target_id", "scene", "risk_level", "hit_reason", "request_id", "created_at"}).
			AddRow(int64(11), int64(0), "user", int64(1001), "login", int32(2), "异常登录", "req-1", createdAt))
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM admin_operation_log WHERE module = 'risk' AND action = 'review_pass' AND target_type = 'risk_hit_record' AND target_id = \?`).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	resp, err := NewHandleRiskHitRecordsLogic(context.Background(), svcCtx).HandleRiskHitRecords(&adminsvc.RiskHitActionRequest{
		Ids: []int64{11}, Action: "review_pass", AdminId: 9001,
	})
	if err != nil {
		t.Fatalf("HandleRiskHitRecords() error = %v", err)
	}
	if resp.GetSuccessCount() != 0 || resp.GetFailCount() != 1 || len(resp.GetFailureReasons()) != 1 {
		t.Fatalf("HandleRiskHitRecords() = %#v, want one duplicate failure", resp)
	}
}

// TestCreateExportTask_RejectsUnknownFilter 验证创建任务时拒绝未定义筛选字段，防止条件被静默忽略。
func TestCreateExportTask_RejectsUnknownFilter(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	_, err := NewCreateExportTaskLogic(context.Background(), svcCtx).CreateExportTask(&adminsvc.ExportTaskRequest{
		ExportType: "orders", Filters: `{"unknown":1}`, AdminId: 9001,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("CreateExportTask() error code = %v, want InvalidArgument", status.Code(err))
	}
}

// TestPromotionActionValidation 验证发布范围和目标配置必须符合活动状态机入口约束。
func TestPromotionActionValidation(t *testing.T) {
	err := validatePromotionAction(&adminsvc.PromotionActivityActionRequest{Id: 1, AdminId: 1, PublishScope: "all", TargetConfig: `{}`}, true)
	if err != nil {
		t.Fatalf("validatePromotionAction() error = %v, want all scope accepted", err)
	}
	err = validatePromotionAction(&adminsvc.PromotionActivityActionRequest{Id: 1, AdminId: 1, PublishScope: "gray", TargetConfig: `{}`}, true)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("validatePromotionAction(gray empty) code = %v, want InvalidArgument", status.Code(err))
	}
}

// TestPromotionActivityValidation_RequiresTypedRule 验证活动配置必须按活动类型提供真实规则字段。
func TestPromotionActivityValidation_RequiresTypedRule(t *testing.T) {
	err := validatePromotionActivity(&adminsvc.PromotionActivityRequest{
		Name: "折扣活动", Type: 2, Config: `{"discount":"0.85"}`, StartAt: "2026-08-20 00:00:00", EndAt: "2026-08-21 00:00:00", Status: 1, AdminId: 9001,
	})
	if err != nil {
		t.Fatalf("validatePromotionActivity() error = %v, want valid discount activity", err)
	}
	err = validatePromotionActivity(&adminsvc.PromotionActivityRequest{
		Name: "空活动", Type: 2, Config: `{}`, StartAt: "2026-08-20 00:00:00", EndAt: "2026-08-21 00:00:00", Status: 1, AdminId: 9001,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("validatePromotionActivity(empty rule) code = %v, want InvalidArgument", status.Code(err))
	}
}

// TestListPromotionActivities_ReturnsPublishState 验证活动列表会回填最近一次发布范围、目标配置和回滚时间。
// 这些字段来自结构化操作日志，不修改 promotion_activity 表结构，便于运营后台确认投放闭环。
func TestListPromotionActivities_ReturnsPublishState(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	startAt := time.Date(2026, 8, 28, 9, 0, 0, 0, time.Local)
	endAt := time.Date(2026, 8, 29, 9, 0, 0, 0, time.Local)
	createdAt := time.Date(2026, 8, 28, 8, 0, 0, 0, time.Local)
	publishedAt := time.Date(2026, 8, 28, 10, 0, 0, 0, time.Local)
	rollbackAt := time.Date(2026, 8, 28, 11, 0, 0, 0, time.Local)
	detail := `{"action":"publish","publish_scope":"gray","target_config":"{\"user_ids\":[1001]}"}`

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM promotion_activity p WHERE 1=1 AND p\.status = \?`).
		WithArgs(int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT p\.id, p\.name, p\.type, p\.config, p\.start_at, p\.end_at, p\.status, p\.created_by, p\.created_at, p\.updated_at,`).
		WithArgs(int32(2), int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "type", "config", "start_at", "end_at", "status", "created_by", "created_at", "updated_at",
			"publish_detail", "published_at", "rollback_at",
		}).AddRow(int64(7), "灰度折扣活动", int32(2), `{"discount":"0.85"}`, startAt, endAt, int32(2), int64(9001), createdAt, createdAt, detail, publishedAt, rollbackAt))

	resp, err := NewListPromotionActivitiesLogic(context.Background(), svcCtx).ListPromotionActivities(&adminsvc.PromotionActivityListRequest{
		Page: 1, PageSize: 20, Status: 2,
	})
	if err != nil {
		t.Fatalf("ListPromotionActivities() error = %v", err)
	}
	if len(resp.GetList()) != 1 {
		t.Fatalf("ListPromotionActivities() list len = %d, want 1", len(resp.GetList()))
	}
	item := resp.GetList()[0]
	if item.GetPublishScope() != "gray" || item.GetTargetConfig() != `{"user_ids":[1001]}` || item.GetPublishedAt() == "" || item.GetRollbackAt() == "" {
		t.Fatalf("PromotionActivity publish state = %#v, want structured publish state", item)
	}
}

// TestPromotionActionDetail_RoundTrip 验证活动动作日志使用可解析 JSON，支撑列表回填投放目标。
func TestPromotionActionDetail_RoundTrip(t *testing.T) {
	item := &adminsvc.PromotionActivity{}
	applyPromotionPublishState(item, promotionActionDetail("publish", "all", `{}`), sql.NullTime{}, sql.NullTime{})
	if item.GetPublishScope() != "all" || item.GetTargetConfig() != `{}` {
		t.Fatalf("applyPromotionPublishState() = %#v, want parsed publish detail", item)
	}
}

// TestGetStatisticsOverview_RejectsUnsupportedCityFilter 验证不存在城市归属字段时拒绝城市筛选，避免静默返回跨城市口径。
func TestGetStatisticsOverview_RejectsUnsupportedCityFilter(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()
	_, err := NewGetStatisticsOverviewLogic(context.Background(), svcCtx).GetStatisticsOverview(&adminsvc.StatisticsRequest{CityCode: "110000"})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("GetStatisticsOverview() error code = %v, want FailedPrecondition", status.Code(err))
	}
}

// TestIssueCoupon_RejectsDraftCoupon 验证草稿券不能进入发券事务，防止未发布规则提前影响用户权益。
func TestIssueCoupon_RejectsDraftCoupon(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT valid_end_at, status, total_count, received_count, per_user_limit`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"valid_end_at", "status", "total_count", "received_count", "per_user_limit"}).
			AddRow(time.Now().Add(time.Hour), int32(1), int64(10), int64(0), int64(1)))
	mock.ExpectRollback()
	_, err := NewIssueCouponLogic(context.Background(), svcCtx).IssueCoupon(&adminsvc.CouponIssueRequest{
		CouponId: 10, AdminId: 1, TargetType: "user", TargetConfig: `{"user_ids":[1001]}`,
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("IssueCoupon() error code = %v, want FailedPrecondition", status.Code(err))
	}
}

// TestIssueCoupon_WritesPublishRecordInTransaction 验证实际发券、任务和发布记录必须同事务提交。
func TestIssueCoupon_WritesPublishRecordInTransaction(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	validEndAt := time.Now().Add(time.Hour)
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT valid_end_at, status, total_count, received_count, per_user_limit`).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"valid_end_at", "status", "total_count", "received_count", "per_user_limit"}).
			AddRow(validEndAt, int32(2), int64(10), int64(0), int64(1)))
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM user_coupon WHERE user_id = \? AND coupon_id = \?`).
		WithArgs(int64(1001), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec(`INSERT INTO user_coupon`).
		WithArgs(int64(1001), int64(10), sqlmock.AnyArg(), validEndAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO admin_coupon_issue_task`).
		WithArgs(sqlmock.AnyArg(), int64(10), "user", `{"user_ids":[1001]}`, 1, int64(1), int64(0), int32(3), "", int64(9001)).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec(`UPDATE coupon SET received_count = received_count \+ \? WHERE id = \?`).
		WithArgs(int64(1), int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO admin_operation_log`).
		WithArgs(int64(9001), "coupon", "issue", "coupon", int64(10), sqlmock.AnyArg(), "127.0.0.1").
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec(`INSERT INTO admin_coupon_publish_record`).
		WithArgs(int64(10), sqlmock.AnyArg(), `{"user_ids":[1001]}`, int32(2), "", int64(9001)).
		WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectCommit()

	resp, err := NewIssueCouponLogic(context.Background(), svcCtx).IssueCoupon(&adminsvc.CouponIssueRequest{
		CouponId: 10, AdminId: 9001, TargetType: "user", TargetConfig: `{"user_ids":[1001]}`, Ip: "127.0.0.1",
	})
	if err != nil || resp.GetStatus() != "success" {
		t.Fatalf("IssueCoupon() = %#v, %v; want successful transaction", resp, err)
	}
}

// TestGetCouponStatistics_UsesSingleAggregateQuery 验证优惠券统计使用一次聚合查询，并且启用券只统计状态 2。
func TestGetCouponStatistics_CountsEnabledStatus(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT\s+\(SELECT COUNT\(1\) FROM coupon\),`).
		WillReturnRows(sqlmock.NewRows([]string{"coupon_count", "enabled_coupon_count", "issued_coupon_count", "used_coupon_count", "expired_coupon_count"}).
			AddRow(3, 1, 0, 0, 0))
	resp, err := NewGetCouponStatisticsLogic(context.Background(), svcCtx).GetCouponStatistics(&adminsvc.StatisticsRequest{})
	if err != nil || resp.GetEnabledCouponCount() != 1 {
		t.Fatalf("GetCouponStatistics() = %#v, %v; want one enabled coupon", resp, err)
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
