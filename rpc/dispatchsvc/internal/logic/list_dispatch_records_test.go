package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"
)

// TestListDispatchRecordsByDriver 验证按司机分页查询派单记录：跨订单汇总该司机全部记录。
func TestListDispatchRecordsByDriver(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	if _, err := NewDispatchOrderLogic(ctx, svcCtx).DispatchOrder(&proto.DispatchOrderRequest{OrderId: 1}); err != nil {
		t.Fatalf("seed order 1 error = %v", err)
	}
	if _, err := NewDispatchOrderLogic(ctx, svcCtx).DispatchOrder(&proto.DispatchOrderRequest{OrderId: 2}); err != nil {
		t.Fatalf("seed order 2 error = %v", err)
	}

	l := NewListDispatchRecordsLogic(ctx, svcCtx)

	// 司机 9001 在两个订单中各有一条派单记录。
	resp, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{DriverId: 9001, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDispatchRecords() by driver error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 2 {
		t.Fatalf("ListDispatchRecords() by driver total = %d len = %d, want 2/2", resp.Total, len(resp.List))
	}
	for _, item := range resp.List {
		if item.DriverId != 9001 {
			t.Fatalf("ListDispatchRecords() item driverId = %d, want 9001", item.DriverId)
		}
	}

	// 分页：page_size=1 只回一条，total 仍为 2。
	first, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{DriverId: 9001, Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("ListDispatchRecords() by driver page error = %v", err)
	}
	if first.Total != 2 || len(first.List) != 1 {
		t.Fatalf("ListDispatchRecords() by driver page total = %d len = %d, want 2/1", first.Total, len(first.List))
	}

	// 未派单过的司机返回空列表。
	empty, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{DriverId: 9999, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDispatchRecords() empty error = %v", err)
	}
	if empty.Total != 0 || len(empty.List) != 0 {
		t.Fatalf("ListDispatchRecords() empty total = %d len = %d, want 0/0", empty.Total, len(empty.List))
	}
}

// TestListDispatchRecordsByDriverWithStatusFilter 验证按司机 + 状态过滤：拒单后按 REJECTED 能查到该记录。
func TestListDispatchRecordsByDriverWithStatusFilter(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	if _, err := NewDispatchOrderLogic(ctx, svcCtx).DispatchOrder(&proto.DispatchOrderRequest{OrderId: 1}); err != nil {
		t.Fatalf("seed order 1 error = %v", err)
	}
	// 司机 9001 拒单。
	if _, err := NewRejectDispatchLogic(ctx, svcCtx).RejectDispatch(&proto.RejectDispatchRequest{
		OrderId: 1, DriverId: 9001, Reason: "暂时无法接单",
	}); err != nil {
		t.Fatalf("RejectDispatch() error = %v", err)
	}

	l := NewListDispatchRecordsLogic(ctx, svcCtx)

	// 按 REJECTED 过滤：命中 9001。
	rejected, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{
		DriverId: 9001, Status: int32(constants.DispatchStatusRejected), Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDispatchRecords() rejected error = %v", err)
	}
	if rejected.Total != 1 || len(rejected.List) != 1 || rejected.List[0].Status != int32(constants.DispatchStatusRejected) {
		t.Fatalf("ListDispatchRecords() rejected = %+v, want 1 REJECTED record", rejected)
	}

	// 按 PENDING 过滤：9001 已拒，剩 9002/9003 仍是 Pending。
	pending, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{
		DriverId: 9001, Status: int32(constants.DispatchStatusPending), Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDispatchRecords() pending error = %v", err)
	}
	if pending.Total != 0 || len(pending.List) != 0 {
		t.Fatalf("ListDispatchRecords() pending = %+v, want 0 records for 9001", pending)
	}

	// 状态过滤只影响 driver_id 查询，不误伤其他司机。
	other, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{
		DriverId: 9002, Status: int32(constants.DispatchStatusPending), Page: 1, PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListDispatchRecords() other error = %v", err)
	}
	if other.Total != 1 || other.List[0].DriverId != 9002 {
		t.Fatalf("ListDispatchRecords() other = %+v, want 1 PENDING record for 9002", other)
	}
}

// TestListDispatchRecordsRejectsNoFilter 验证 order_id 与 driver_id 均为 0 时拒绝查询。
func TestListDispatchRecordsRejectsNoFilter(t *testing.T) {
	l := NewListDispatchRecordsLogic(context.Background(), newDispatchTestSvcCtx())

	_, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ListDispatchRecords() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}
