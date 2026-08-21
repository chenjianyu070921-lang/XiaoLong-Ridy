package adminservicelogic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
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

// TestListWorkOrders_ReturnsCountQueryError 验证工单列表总数查询失败时不返回不完整结果。
func TestListWorkOrders_ReturnsCountQueryError(t *testing.T) {
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM admin_complaint_work_order`).WillReturnError(errors.New("work order database unavailable"))

	if _, err := NewWorkOrderLogic(context.Background(), svcCtx).ListWorkOrders(&adminsvc.WorkOrderListRequest{}); err == nil {
		t.Fatal("ListWorkOrders() error = nil, want database error")
	}
}
