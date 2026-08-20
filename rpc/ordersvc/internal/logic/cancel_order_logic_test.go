package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"XiaoLong-Ridy/rpc/ordersvc/internal/model"
	"XiaoLong-Ridy/rpc/ordersvc/internal/repository"
	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"
)

var seedOrderSeq uint64

func nextTestOrderNo() string {
	seedOrderSeq++
	return fmt.Sprintf("TEST%d", seedOrderSeq)
}

func seedOrder(t *testing.T, repo *repository.MemoryOrderRepository, userID, driverID uint64, status int8) *model.RideOrder {
	t.Helper()
	order := &model.RideOrder{
		OrderNo:       nextTestOrderNo(),
		UserId:        userID,
		DriverId:      driverID,
		CarType:       1,
		FromAddress:   "起点",
		FromLongitude: 116.1,
		FromLatitude:  39.1,
		ToAddress:     "终点",
		ToLongitude:   116.2,
		ToLatitude:    39.2,
		Status:        status,
	}
	statusLog := &model.OrderStatusLog{
		FromStatus:   0,
		ToStatus:     status,
		OperatorType: "user",
		OperatorId:   userID,
		Remark:       "seed",
	}
	if err := repo.Create(context.Background(), order, statusLog); err != nil {
		t.Fatalf("seed orderclient error = %v", err)
	}
	return order
}

func TestCancelOrderSuccessByUser(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "user",
		OperatorId:   1001,
		Reason:       "行程有变",
	})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Fatalf("CancelOrder() status = %v, want CANCELLED", resp.Status)
	}

	fresh, err := repo.GetByID(context.Background(), order.Id)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if fresh.Status != 6 || fresh.CancelBy != "user" || fresh.CancelReason != "行程有变" {
		t.Fatalf("cancelled orderclient = %+v", fresh)
	}
	logs := repo.StatusLogs(order.Id)
	if len(logs) != 2 || logs[1].FromStatus != 1 || logs[1].ToStatus != 6 || logs[1].OperatorType != "user" {
		t.Fatalf("status logs = %+v", logs)
	}
}

func TestCancelOrderSuccessByDriver(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "driver",
		OperatorId:   2002,
		Reason:       "车辆故障",
	})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Fatalf("CancelOrder() status = %v, want CANCELLED", resp.Status)
	}
}

func TestCancelOrderSuccessBySystem(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "system",
		OperatorId:   0,
		Reason:       "超时未接单",
	})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Fatalf("CancelOrder() status = %v, want CANCELLED", resp.Status)
	}
}

func TestCancelOrderRejectNonOwnerUser(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "user",
		OperatorId:   9999,
		Reason:       "行程有变",
	})
	if !errors.Is(err, ErrCancelNotAllowed) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrCancelNotAllowed)
	}
}

func TestCancelOrderRejectDriverNotAssigned(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 2)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "driver",
		OperatorId:   8888,
		Reason:       "车辆故障",
	})
	if !errors.Is(err, ErrCancelNotAllowed) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrCancelNotAllowed)
	}
}

func TestCancelOrderRejectDriverOnWaitAccept(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "driver",
		OperatorId:   2002,
		Reason:       "不接单",
	})
	if !errors.Is(err, ErrCancelNotAllowed) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrCancelNotAllowed)
	}
}

func TestCancelOrderRejectOnTrip(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 2002, 3)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "user",
		OperatorId:   1001,
		Reason:       "行程有变",
	})
	if !errors.Is(err, ErrOrderStatusNotCancelable) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrOrderStatusNotCancelable)
	}
}

func TestCancelOrderRejectAlreadyCancelled(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 6)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "user",
		OperatorId:   1001,
		Reason:       "重复取消",
	})
	if !errors.Is(err, ErrOrderStatusNotCancelable) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrOrderStatusNotCancelable)
	}
}

func TestCancelOrderRejectEmptyReason(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "user",
		OperatorId:   1001,
		Reason:       " ",
	})
	if !errors.Is(err, ErrCancelReasonRequired) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrCancelReasonRequired)
	}
}

func TestCancelOrderRejectInvalidOperator(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := seedOrder(t, repo, 1001, 0, 1)
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      int64(order.Id),
		OperatorType: "boss",
		OperatorId:   1001,
		Reason:       "行程有变",
	})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}

func TestCancelOrderNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewCancelOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.CancelOrder(&proto.CancelOrderRequest{
		OrderId:      999,
		OperatorType: "system",
		Reason:       "测试",
	})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("CancelOrder() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}
