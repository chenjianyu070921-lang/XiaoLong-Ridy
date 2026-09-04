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

// TestRejectDispatchFlow verifies rejecting a dispatch marks only that driver's
// record as rejected while other candidates stay pending.
func TestRejectDispatchFlow(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	seedDispatch(t, svcCtx, 1)

	l := NewRejectDispatchLogic(ctx, svcCtx)
	resp, err := l.RejectDispatch(&proto.RejectDispatchRequest{OrderId: 1, DriverId: 9001, Reason: "too far"})
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
	reasonByDriver := make(map[uint64]string)
	for _, r := range records {
		statusByDriver[r.DriverId] = int32(r.Status)
		reasonByDriver[r.DriverId] = r.RejectReason
	}
	if statusByDriver[9001] != constants.DispatchStatusRejected {
		t.Fatalf("driver 9001 status = %d, want Rejected", statusByDriver[9001])
	}
	if reasonByDriver[9001] != "too far" {
		t.Fatalf("driver 9001 reject reason = %q, want too far", reasonByDriver[9001])
	}

	if statusByDriver[9002] != constants.DispatchStatusPending || statusByDriver[9003] != constants.DispatchStatusPending {
		t.Fatalf("other candidates should stay Pending, got %+v", statusByDriver)
	}

	// Repeated rejection should fail because the record is no longer pending.
	if _, err := l.RejectDispatch(&proto.RejectDispatchRequest{OrderId: 1, DriverId: 9001}); !errors.Is(err, repository.ErrDispatchRecordNotFound) {
		t.Fatalf("RejectDispatch() second error = %v, want ErrDispatchRecordNotFound", err)
	}
}

// TestCancelDispatchFlow verifies order cancellation marks all pending dispatch
// records as cancelled.
func TestCancelDispatchFlow(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	seedDispatch(t, svcCtx, 2)

	l := NewCancelDispatchLogic(ctx, svcCtx)
	resp, err := l.CancelDispatch(&proto.CancelDispatchRequest{OrderId: 2, Reason: "too far"})
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

	// Repeated cancellation is idempotent for already-cancelled records.
	again, err := l.CancelDispatch(&proto.CancelDispatchRequest{OrderId: 2})
	if err != nil {
		t.Fatalf("CancelDispatch() again error = %v", err)
	}
	if again.Affected != 0 {
		t.Fatalf("CancelDispatch() again affected = %d, want 0", again.Affected)
	}
}

// TestListTimeoutPendingOrdersNoTimeout verifies fresh pending dispatches are
// not returned by timeout scanning.
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
