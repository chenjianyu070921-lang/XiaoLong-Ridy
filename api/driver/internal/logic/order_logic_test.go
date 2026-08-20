package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

type fakeOrderClient struct {
	acceptRequest *orderproto.AcceptOrderRequest
	arriveRequest *orderproto.ConfirmArriveRequest
	startRequest  *orderproto.StartTripRequest
	finishRequest *orderproto.FinishTripRequest
}

func (f *fakeOrderClient) GetOrder(_ context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	return &orderproto.GetOrderResponse{OrderId: req.OrderId, OrderNo: "NO-1001", FromAddress: "起点", ToAddress: "终点", Status: orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT, EstimatedPriceCents: 29900, CreatedAt: 100}, nil
}

func (f *fakeOrderClient) AcceptOrder(_ context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	f.acceptRequest = req
	return &orderproto.AcceptOrderResponse{OrderId: req.OrderId, Status: orderproto.OrderStatus_ORDER_STATUS_ACCEPTED}, nil
}

func (f *fakeOrderClient) StartTrip(_ context.Context, req *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error) {
	f.startRequest = req
	return &orderproto.StartTripResponse{OrderId: req.OrderId, Status: orderproto.OrderStatus_ORDER_STATUS_ON_TRIP}, nil
}

func (f *fakeOrderClient) ConfirmArrive(_ context.Context, req *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error) {
	f.arriveRequest = req
	return &orderproto.ConfirmArriveResponse{OrderId: req.OrderId, Status: orderproto.OrderStatus_ORDER_STATUS_ACCEPTED}, nil
}

func (f *fakeOrderClient) FinishTrip(_ context.Context, req *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error) {
	f.finishRequest = req
	return &orderproto.FinishTripResponse{
		OrderId:            req.OrderId,
		Status:             orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY,
		PayableAmountCents: req.ActualPriceCents,
	}, nil
}

func TestAcceptOrderForwardsDriverAndOrder(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.AcceptOrder(25, 1001)
	if err != nil {
		t.Fatalf("AcceptOrder() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_ACCEPTED) {
		t.Fatalf("AcceptOrder() response = %+v", resp)
	}
	if client.acceptRequest.GetDriverId() != 25 || client.acceptRequest.GetOrderId() != 1001 {
		t.Fatalf("AcceptOrder() request = %+v", client.acceptRequest)
	}
}

func TestConfirmArriveForwardsDriverAndOrder(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.ConfirmArrive(25, 1001)
	if err != nil {
		t.Fatalf("ConfirmArrive() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_ACCEPTED) {
		t.Fatalf("ConfirmArrive() response = %+v", resp)
	}
	if client.arriveRequest.GetDriverId() != 25 || client.arriveRequest.GetOrderId() != 1001 {
		t.Fatalf("ConfirmArrive() request = %+v", client.arriveRequest)
	}
}

func TestStartTripForwardsDriverAndOrder(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.StartTrip(25, 1001)
	if err != nil {
		t.Fatalf("StartTrip() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_ON_TRIP) {
		t.Fatalf("StartTrip() response = %+v", resp)
	}
	if client.startRequest.GetDriverId() != 25 || client.startRequest.GetOrderId() != 1001 {
		t.Fatalf("StartTrip() request = %+v", client.startRequest)
	}
}

func TestFinishTripForwardsTripMetrics(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})
	req := &types.FinishTripRequest{
		OrderID:          1001,
		ActualDistanceM:  12500,
		ActualDurationS:  1800,
		ActualPriceCents: 3200,
	}

	resp, err := logic.FinishTrip(25, req)
	if err != nil {
		t.Fatalf("FinishTrip() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY) || resp.PayableAmountCents != 3200 {
		t.Fatalf("FinishTrip() response = %+v", resp)
	}
	if client.finishRequest.GetDriverId() != 25 || client.finishRequest.GetOrderId() != 1001 ||
		client.finishRequest.GetActualDistanceM() != 12500 || client.finishRequest.GetActualDurationS() != 1800 ||
		client.finishRequest.GetActualPriceCents() != 3200 {
		t.Fatalf("FinishTrip() request = %+v", client.finishRequest)
	}
}

func TestOrderLogicRejectsInvalidOrderParameters(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	if _, err := logic.AcceptOrder(0, 1001); err != ErrInvalidParam {
		t.Fatalf("AcceptOrder() error = %v, want %v", err, ErrInvalidParam)
	}
	if _, err := logic.ConfirmArrive(25, 0); err != ErrInvalidParam {
		t.Fatalf("ConfirmArrive() error = %v, want %v", err, ErrInvalidParam)
	}
	if _, err := logic.StartTrip(-1, 1001); err != ErrInvalidParam {
		t.Fatalf("StartTrip() error = %v, want %v", err, ErrInvalidParam)
	}
	if _, err := logic.FinishTrip(25, nil); err != ErrInvalidParam {
		t.Fatalf("FinishTrip(nil) error = %v, want %v", err, ErrInvalidParam)
	}
	if _, err := logic.FinishTrip(25, &types.FinishTripRequest{OrderID: 1001, ActualDistanceM: -1}); err != ErrInvalidParam {
		t.Fatalf("FinishTrip(negative distance) error = %v, want %v", err, ErrInvalidParam)
	}
}

func TestRejectOrderRequiresDispatchClient(t *testing.T) {
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: &fakeDispatchClient{}})

	_, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001})
	if err != nil {
		t.Fatalf("RejectOrder() error = %v", err)
	}
}

func TestListMyDispatchesRequiresOrderClient(t *testing.T) {
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: &fakeDispatchClient{}})

	_, err := logic.ListMyDispatches(25, 0, 1000, 0)
	if err != ErrOrderClientNotConfigured {
		t.Fatalf("ListMyDispatches() error = %v, want %v", err, ErrOrderClientNotConfigured)
	}
}

type fakeDispatchClient struct {
	rejectRequest *dispatchproto.RejectDispatchRequest
	listRequest   *dispatchproto.ListDispatchRecordsRequest
}

func (f *fakeDispatchClient) RejectDispatch(_ context.Context, req *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error) {
	f.rejectRequest = req
	return &dispatchproto.RejectDispatchResponse{OrderId: req.OrderId, DriverId: req.DriverId, Status: 2}, nil
}

func (f *fakeDispatchClient) ListDispatchRecords(_ context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	f.listRequest = req
	return &dispatchproto.ListDispatchRecordsResponse{List: []*dispatchproto.DispatchRecord{{Id: 7, OrderId: 1001, DriverId: req.DriverId, Status: 1}}, Total: 1, Page: req.Page, PageSize: req.PageSize}, nil
}

func TestRejectOrderForwardsDispatchRequest(t *testing.T) {
	client := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

	resp, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "临时有事"})
	if err != nil || resp.OrderID != 1001 || resp.DriverID != 25 {
		t.Fatalf("RejectOrder() response = %+v, error = %v", resp, err)
	}
	if client.rejectRequest.GetOrderId() != 1001 || client.rejectRequest.GetDriverId() != 25 || client.rejectRequest.GetReason() != "临时有事" {
		t.Fatalf("RejectOrder() request = %+v", client.rejectRequest)
	}
}

func TestListMyDispatchesCombinesDispatchAndOrder(t *testing.T) {
	dispatchClient := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: dispatchClient, OrderClient: &fakeOrderClient{}})

	resp, err := logic.ListMyDispatches(25, 1, 20, 1)
	if err != nil || len(resp.List) != 1 || resp.List[0].Order.OrderNo != "NO-1001" {
		t.Fatalf("ListMyDispatches() response = %+v, error = %v", resp, err)
	}
	if dispatchClient.listRequest.GetDriverId() != 25 || dispatchClient.listRequest.GetStatus() != 1 {
		t.Fatalf("ListMyDispatches() request = %+v", dispatchClient.listRequest)
	}
}
