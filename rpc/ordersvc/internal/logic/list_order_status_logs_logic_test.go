package logic

import (
	"context"
	"errors"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

func TestListOrderStatusLogsPagination(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	if err := repo.AppendStatusLog(context.Background(), &model.OrderStatusLog{
		OrderId:      order.Id,
		FromStatus:   1,
		ToStatus:     2,
		OperatorType: "driver",
		OperatorId:   2002,
		Remark:       "司机接单",
	}); err != nil {
		t.Fatalf("AppendStatusLog() error = %v", err)
	}
	l := NewListOrderStatusLogsLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	first, err := l.ListOrderStatusLogs(&proto.ListOrderStatusLogsRequest{
		OrderId:  int64(order.Id),
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListOrderStatusLogs() error = %v", err)
	}
	if first.Total != 2 || len(first.List) != 1 || first.Page != 1 || first.PageSize != 1 {
		t.Fatalf("ListOrderStatusLogs() response = %+v", first)
	}
	if first.List[0].FromStatus != 0 || first.List[0].ToStatus != 1 {
		t.Fatalf("first log = %+v", first.List[0])
	}

	second, err := l.ListOrderStatusLogs(&proto.ListOrderStatusLogsRequest{
		OrderId:  int64(order.Id),
		Page:     2,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListOrderStatusLogs() error = %v", err)
	}
	if len(second.List) != 1 || second.List[0].FromStatus != 1 || second.List[0].ToStatus != 2 {
		t.Fatalf("second log = %+v", second.List[0])
	}
}

func TestListOrderStatusLogsEmpty(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewListOrderStatusLogsLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.ListOrderStatusLogs(&proto.ListOrderStatusLogsRequest{OrderId: 999})
	if err != nil {
		t.Fatalf("ListOrderStatusLogs() error = %v", err)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListOrderStatusLogs() response = %+v", resp)
	}
}

func TestListOrderStatusLogsRejectInvalidID(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewListOrderStatusLogsLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.ListOrderStatusLogs(&proto.ListOrderStatusLogsRequest{OrderId: 0})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("ListOrderStatusLogs() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}
