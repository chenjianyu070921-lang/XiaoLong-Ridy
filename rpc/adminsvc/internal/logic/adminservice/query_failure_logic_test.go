package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"

	"github.com/DATA-DOG/go-sqlmock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestListUsers_ReturnsCountQueryError 验证用户列表的分页总数查询失败时不会返回空列表假成功。
func TestListUsers_ReturnsCountQueryError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM user`).WillReturnError(errors.New("users database unavailable"))

	if _, err := NewListUsersLogic(context.Background(), svcCtx).ListUsers(&adminsvc.UserListRequest{}); err == nil {
		t.Fatal("ListUsers() error = nil, want database error")
	}
}

// TestListOrders_ReturnsCountQueryError 验证订单列表的分页总数查询失败时不会继续读取订单明细。
func TestListOrders_ReturnsCountQueryError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM ride_order`).WillReturnError(errors.New("orders database unavailable"))

	if _, err := NewListOrdersLogic(context.Background(), svcCtx).ListOrders(&adminsvc.OrderListRequest{}); err == nil {
		t.Fatal("ListOrders() error = nil, want database error")
	}
}

// TestGetOrderStatistics_ReturnsAggregateQueryError 验证订单统计聚合查询失败时向调用方返回错误。
func TestGetOrderStatistics_ReturnsAggregateQueryError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT\s+\(SELECT COUNT\(1\) FROM ride_order`).WillReturnError(errors.New("statistics database unavailable"))

	if _, err := NewGetOrderStatisticsLogic(context.Background(), svcCtx).GetOrderStatistics(&adminsvc.StatisticsRequest{}); err == nil {
		t.Fatal("GetOrderStatistics() error = nil, want database error")
	}
}

// TestGetOrderStatistics_AppliesOrderTimeRangeToRelatedRecords 验证超时和支付异常统计按订单创建时间过滤。
func TestGetOrderStatistics_AppliesOrderTimeRangeToRelatedRecords(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	startTime := "2026-08-01 00:00:00"
	endTime := "2026-08-24 23:59:59"
	mock.ExpectQuery(`(?s)SELECT\s+\(SELECT COUNT\(1\) FROM ride_order ro WHERE ro\.created_at >= \? AND ro\.created_at <= \?\).*JOIN ride_order ro ON ro\.id = dr\.order_id.*JOIN ride_order ro ON ro\.id = p\.order_id`).
		WithArgs(
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"order_count", "completed_order_count", "canceled_order_count",
			"timeout_order_count", "payment_abnormal_count",
		}).AddRow(100, 70, 20, 6, 4))

	resp, err := NewGetOrderStatisticsLogic(context.Background(), svcCtx).GetOrderStatistics(&adminsvc.StatisticsRequest{
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("GetOrderStatistics() error = %v", err)
	}
	if resp.GetOrderCount() != 100 || resp.GetCompletedOrderCount() != 70 ||
		resp.GetCanceledOrderCount() != 20 || resp.GetTimeoutOrderCount() != 6 ||
		resp.GetPaymentAbnormalCount() != 4 {
		t.Fatalf("GetOrderStatistics() response = %+v, want order counts 100/70/20/6/4", resp)
	}
	if resp.GetCompletionRate() != "70.00%" || resp.GetCancelRate() != "20.00%" {
		t.Fatalf("GetOrderStatistics() rates = %q/%q, want 70.00%%/20.00%%",
			resp.GetCompletionRate(), resp.GetCancelRate())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestGetDriverStatistics_AppliesTimeRangeToDriverReportSources 验证司机报表按各来源表时间字段过滤。
// 司机在线状态和独立接单事件当前没有可靠数据表，本用例只断言已开放的真实统计指标。
func TestGetDriverStatistics_AppliesTimeRangeToDriverReportSources(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	startTime := "2026-08-01 00:00:00"
	endTime := "2026-08-24 23:59:59"
	mock.ExpectQuery(`(?s)SELECT\s+\(SELECT COUNT\(1\) FROM driver WHERE deleted_at IS NULL\).*FROM driver d WHERE d\.created_at >= \? AND d\.created_at <= \? AND d\.deleted_at IS NULL.*FROM driver_certification dc WHERE dc\.created_at >= \? AND dc\.created_at <= \? AND dc\.audit_status = 1.*FROM settlement s WHERE s\.created_at >= \? AND s\.created_at <= \? AND s\.status = 2.*FROM driver_withdraw dw WHERE dw\.created_at >= \? AND dw\.created_at <= \? AND dw\.status = 1.*FROM driver_withdraw dw WHERE dw\.created_at >= \? AND dw\.created_at <= \? AND dw\.status = 3.*FROM driver_withdraw dw WHERE dw\.created_at >= \? AND dw\.created_at <= \? AND dw\.status = 4.*FROM driver_score ds WHERE ds\.updated_at >= \? AND ds\.updated_at <= \?`).
		WithArgs(
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"driver_total", "new_driver_count", "pending_audit_count", "approved_driver_count",
			"completed_order_count", "driver_income", "withdraw_pending_amount",
			"withdraw_success_amount", "withdraw_failed_count", "average_score", "total_complaint_count",
		}).AddRow(20, 3, 2, 4, 16, "888.00", "120.00", "560.00", 1, "96.50", 5))

	resp, err := NewGetDriverStatisticsLogic(context.Background(), svcCtx).GetDriverStatistics(&adminsvc.StatisticsRequest{
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("GetDriverStatistics() error = %v", err)
	}
	if resp.GetNewDriverCount() != 3 || resp.GetPendingAuditCount() != 2 ||
		resp.GetCompletedOrderCount() != 16 || resp.GetDriverIncome() != "888.00" ||
		resp.GetAverageScore() != "96.50" {
		t.Fatalf("GetDriverStatistics() response = %+v, want driver report metrics", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestGetFinanceStatistics_AppliesTimeRangeToFinanceSources 验证财务报表按支付、结算和价格明细各自创建时间过滤。
func TestGetFinanceStatistics_AppliesTimeRangeToFinanceSources(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()

	startTime := "2026-08-01 00:00:00"
	endTime := "2026-08-24 23:59:59"
	mock.ExpectQuery(`(?s)SELECT\s+\(SELECT COUNT\(1\) FROM payment p WHERE p\.created_at >= \? AND p\.created_at <= \? AND p\.status = 2\).*FROM settlement s WHERE s\.created_at >= \? AND s\.created_at <= \? AND s\.status = 2.*FROM order_price op WHERE op\.created_at >= \? AND op\.created_at <= \?`).
		WithArgs(
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
			startTime, endTime,
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"payment_order_count", "paid_amount", "refund_order_count", "refund_amount",
			"payment_failed_count", "settlement_order_count", "settlement_total_amount",
			"platform_commission", "driver_income", "platform_subsidy",
		}).AddRow(9, "320.00", 2, "30.00", 1, 8, "300.00", "45.00", "255.00", "18.00"))

	resp, err := NewGetFinanceStatisticsLogic(context.Background(), svcCtx).GetFinanceStatistics(&adminsvc.StatisticsRequest{
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		t.Fatalf("GetFinanceStatistics() error = %v", err)
	}
	if resp.GetPaymentOrderCount() != 9 || resp.GetPaidAmount() != "320.00" ||
		resp.GetRefundAmount() != "30.00" || resp.GetPlatformCommission() != "45.00" ||
		resp.GetPlatformSubsidy() != "18.00" {
		t.Fatalf("GetFinanceStatistics() response = %+v, want finance report metrics", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

// TestGetDriverAndFinanceStatistics_RejectUnsupportedCityFilter 验证新增报表沿用城市字段不支持时直接拒绝的口径。
func TestGetDriverAndFinanceStatistics_RejectUnsupportedCityFilter(t *testing.T) {
	svcCtx, _, cleanup := newAdminSQLMock(t)
	defer cleanup()

	driverReq := &adminsvc.StatisticsRequest{CityCode: "110000"}
	if _, err := NewGetDriverStatisticsLogic(context.Background(), svcCtx).GetDriverStatistics(driverReq); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("GetDriverStatistics() error code = %v, want FailedPrecondition", status.Code(err))
	}
	financeReq := &adminsvc.StatisticsRequest{CityCode: "110000"}
	if _, err := NewGetFinanceStatisticsLogic(context.Background(), svcCtx).GetFinanceStatistics(financeReq); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("GetFinanceStatistics() error code = %v, want FailedPrecondition", status.Code(err))
	}
}

// TestListWorkOrders_ReturnsCountQueryError 验证工单列表总数查询失败时不返回不完整结果。
func TestListWorkOrders_ReturnsCountQueryError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM admin_complaint_work_order`).WillReturnError(errors.New("work order database unavailable"))

	if _, err := NewWorkOrderLogic(context.Background(), svcCtx).ListWorkOrders(&adminsvc.WorkOrderListRequest{}); err == nil {
		t.Fatal("ListWorkOrders() error = nil, want database error")
	}
}
