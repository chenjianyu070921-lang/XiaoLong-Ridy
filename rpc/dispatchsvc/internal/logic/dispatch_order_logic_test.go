package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/engine"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/repository"
	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"
)

func newDispatchTestSvcCtx() *svc.ServiceContext {
	return &svc.ServiceContext{
		DispatchRepository: repository.NewMemoryDispatchRepository(),
		DispatchEngine:     engine.NewMockDispatchEngine(),
	}
}

func TestDispatchOrderSuccess(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	l := NewDispatchOrderLogic(ctx, svcCtx)

	resp, err := l.DispatchOrder(&proto.DispatchOrderRequest{
		OrderId:       1,
		FromLongitude: 116.47,
		FromLatitude:  39.9,
		CarType:       1,
		CityCode:      "110000",
	})
	if err != nil {
		t.Fatalf("DispatchOrder() error = %v", err)
	}
	if resp.OrderId != 1 || len(resp.List) != 3 {
		t.Fatalf("DispatchOrder() response = %+v", resp)
	}
	if resp.List[0].DriverId != 9001 || resp.List[0].Status != 1 || resp.List[0].DispatchType != 1 {
		t.Fatalf("DispatchOrder() first record = %+v", resp.List[0])
	}
}

func TestDispatchOrderRejectsInvalidOrderID(t *testing.T) {
	l := NewDispatchOrderLogic(context.Background(), newDispatchTestSvcCtx())

	_, err := l.DispatchOrder(&proto.DispatchOrderRequest{OrderId: 0})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("DispatchOrder() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

func TestListDispatchRecordsPagination(t *testing.T) {
	ctx := context.Background()
	svcCtx := newDispatchTestSvcCtx()
	if _, err := NewDispatchOrderLogic(ctx, svcCtx).DispatchOrder(&proto.DispatchOrderRequest{OrderId: 1}); err != nil {
		t.Fatalf("seed dispatch error = %v", err)
	}
	l := NewListDispatchRecordsLogic(ctx, svcCtx)

	first, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{OrderId: 1, Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListDispatchRecords() error = %v", err)
	}
	if first.Total != 3 || len(first.List) != 2 {
		t.Fatalf("ListDispatchRecords() first page = %+v", first)
	}

	second, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{OrderId: 1, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListDispatchRecords() error = %v", err)
	}
	if second.Total != 3 || len(second.List) != 1 {
		t.Fatalf("ListDispatchRecords() second page = %+v", second)
	}
}

func TestListDispatchRecordsRejectsInvalidOrderID(t *testing.T) {
	l := NewListDispatchRecordsLogic(context.Background(), newDispatchTestSvcCtx())

	_, err := l.ListDispatchRecords(&proto.ListDispatchRecordsRequest{OrderId: 0})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ListDispatchRecords() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}
