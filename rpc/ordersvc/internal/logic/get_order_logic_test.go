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

func TestGetOrderSuccess(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	order := &model.RideOrder{
		OrderNo:            nextTestOrderNo(),
		UserId:             1001,
		DriverId:           2002,
		CarType:            2,
		FromAddress:        "北京市朝阳区建国门外大街1号",
		FromLongitude:      116.46203,
		FromLatitude:       39.90759,
		ToAddress:          "北京市海淀区中关村大街27号",
		ToLongitude:        116.31683,
		ToLatitude:         39.98472,
		EstimatedDistanceM: 15000,
		EstimatedDurationS: 2400,
		EstimatedPrice:     48.5,
		Status:             2,
		CityCode:           "310000",
	}
	statusLog := &model.OrderStatusLog{FromStatus: 0, ToStatus: 2, OperatorType: "user", OperatorId: 1001}
	if err := repo.Create(context.Background(), order, statusLog); err != nil {
		t.Fatalf("seed orderclient error = %v", err)
	}
	l := NewGetOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	resp, err := l.GetOrder(&proto.GetOrderRequest{OrderId: int64(order.Id)})
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if resp.OrderId != int64(order.Id) || resp.UserId != 1001 || resp.DriverId != 2002 {
		t.Fatalf("GetOrder() identity = %+v", resp)
	}
	if resp.CarType != 2 || resp.EstimatedDistanceM != 15000 || resp.EstimatedDurationS != 2400 {
		t.Fatalf("GetOrder() orderclient fields = %+v", resp)
	}
	if resp.CityCode != "310000" {
		t.Fatalf("GetOrder() city code = %q, want 310000", resp.CityCode)
	}
	if resp.EstimatedPriceCents != 4850 {
		t.Fatalf("GetOrder() price = %d, want 4850", resp.EstimatedPriceCents)
	}
	if resp.Status != proto.OrderStatus_ORDER_STATUS_ACCEPTED {
		t.Fatalf("GetOrder() status = %v, want ACCEPTED", resp.Status)
	}
	if resp.CreatedAt <= 0 || resp.UpdatedAt <= 0 {
		t.Fatalf("GetOrder() timestamps = %+v", resp)
	}
}

func TestGetOrderNotFound(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewGetOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.GetOrder(&proto.GetOrderRequest{OrderId: 999})
	if !errors.Is(err, repository.ErrOrderNotFound) {
		t.Fatalf("GetOrder() error = %v, want %v", err, repository.ErrOrderNotFound)
	}
}

func TestGetOrderRejectInvalidID(t *testing.T) {
	repo := repository.NewMemoryOrderRepository()
	l := NewGetOrderLogic(context.Background(), &svc.ServiceContext{OrderRepository: repo})

	_, err := l.GetOrder(&proto.GetOrderRequest{OrderId: 0})
	if !errors.Is(err, ErrInvalidOrderParams) {
		t.Fatalf("GetOrder() error = %v, want %v", err, ErrInvalidOrderParams)
	}
}
