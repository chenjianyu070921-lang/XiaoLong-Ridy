package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/repository"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"
)

func seedDispatch(t *testing.T, svcCtx *svc.ServiceContext, orderID uint64) {
	t.Helper()
	if _, err := NewDispatchOrderLogic(context.Background(), svcCtx).DispatchOrder(&proto.DispatchOrderRequest{
		OrderId: int64(orderID), FromLongitude: 116.47, FromLatitude: 39.9, CarType: 1, CityCode: "110000",
	}); err != nil {
		t.Fatalf("seed dispatch order %d error = %v", orderID, err)
	}
}

// TestRejectDispatchFlow 拒单：该司机记录置 Rejected，其他候选仍为 Pending，重复拒单报错。
func TestRejectDispatchFlow(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	seedDispatch(t, svcCtx, 1)

	l := NewRejectDispatchLogic(ctx, svcCtx)
	resp, err := l.RejectDispatch(&proto.RejectDispatchRequest{OrderId: 1, DriverId: 9001, Reason: "距离太远"})
	if err != nil {
		t.Fatalf("RejectDispatch() error = %v", err)
	}
	if resp.Status != int32(constants.DispatchStatusRejected) {
		t.Fatalf("RejectDispatch() status = %d, want %d", resp.Status, constants.DispatchStatusRejected)
	}

	records, total, err := svcCtx.DispatchRepository.ListByOrder(ctx, 1, 1, 100)
	if err != nil {
		t.Fatalf("ListByOrder() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("ListByOrder() total = %d, want 3", total)
	}
	statusByDriver := make(map[uint64]int32)
	for _, r := range records {
		statusByDriver[r.DriverId] = int32(r.Status)
	}
	if statusByDriver[9001] != constants.DispatchStatusRejected {
		t.Fatalf("driver 9001 status = %d, want Rejected", statusByDriver[9001])
	}
	if statusByDriver[9002] != constants.DispatchStatusPending || statusByDriver[9003] != constants.DispatchStatusPending {
		t.Fatalf("other candidates should stay Pending, got %+v", statusByDriver)
	}

	// 重复拒单：记录已非 Pending，应报记录不存在。
	if _, err := l.RejectDispatch(&proto.RejectDispatchRequest{OrderId: 1, DriverId: 9001}); !errors.Is(err, repository.ErrDispatchRecordNotFound) {
		t.Fatalf("RejectDispatch() second error = %v, want ErrDispatchRecordNotFound", err)
	}
}

// TestCancelDispatchFlow 订单取消：全部 Pending 派单记录同步置为 Cancelled。
func TestCancelDispatchFlow(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	seedDispatch(t, svcCtx, 2)

	l := NewCancelDispatchLogic(ctx, svcCtx)
	resp, err := l.CancelDispatch(&proto.CancelDispatchRequest{OrderId: 2, Reason: "用户取消"})
	if err != nil {
		t.Fatalf("CancelDispatch() error = %v", err)
	}
	if resp.Affected != 3 {
		t.Fatalf("CancelDispatch() affected = %d, want 3", resp.Affected)
	}

	records, total, err := svcCtx.DispatchRepository.ListByOrder(ctx, 2, 1, 100)
	if err != nil {
		t.Fatalf("ListByOrder() error = %v", err)
	}
	if total != 3 {
		t.Fatalf("ListByOrder() total = %d, want 3", total)
	}
	for _, r := range records {
		if r.Status != constants.DispatchStatusCancelled {
			t.Fatalf("record driver=%d status = %d, want Cancelled", r.DriverId, r.Status)
		}
	}

	// 幂等：再次取消不影响已取消记录。
	again, err := l.CancelDispatch(&proto.CancelDispatchRequest{OrderId: 2})
	if err != nil {
		t.Fatalf("CancelDispatch() again error = %v", err)
	}
	if again.Affected != 0 {
		t.Fatalf("CancelDispatch() again affected = %d, want 0", again.Affected)
	}
}

// TestListTimeoutPendingOrdersNoTimeout 新派单未超时时，超时扫描不应返回任何订单。
func TestListTimeoutPendingOrdersNoTimeout(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	seedDispatch(t, svcCtx, 3)

	l := NewListTimeoutPendingOrdersLogic(ctx, svcCtx)
	resp, err := l.ListTimeoutPendingOrders(&proto.ListTimeoutPendingOrdersRequest{TimeoutSeconds: 60, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListTimeoutPendingOrders() error = %v", err)
	}
	if resp.Total != 0 || len(resp.OrderIds) != 0 {
		t.Fatalf("ListTimeoutPendingOrders() = %+v, want empty (records just created)", resp)
	}
}
