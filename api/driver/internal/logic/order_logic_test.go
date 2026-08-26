package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/constants"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

type fakeOrderClient struct {
	getOrderRequest           *orderproto.GetOrderRequest
	getOrderResponseDriverID  int64
	acceptRequest             *orderproto.AcceptOrderRequest
	cancelRequest             *orderproto.CancelOrderRequest
	arriveRequest             *orderproto.ConfirmArriveRequest
	startRequest              *orderproto.StartTripRequest
	finishRequest             *orderproto.FinishTripRequest
	finishResponseAmountCents int64
	listOrdersRequest         *orderproto.ListOrdersRequest
}

func (f *fakeOrderClient) GetOrder(_ context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	f.getOrderRequest = req
	driverID := f.getOrderResponseDriverID
	if driverID == 0 {
		driverID = 25
	}
	return &orderproto.GetOrderResponse{
		OrderId:             req.OrderId,
		OrderNo:             "NO-1001",
		UserId:              300,
		DriverId:            driverID,
		CarType:             1,
		FromAddress:         "璧风偣",
		FromLongitude:       116.391,
		FromLatitude:        39.907,
		ToAddress:           "缁堢偣",
		ToLongitude:         116.481,
		ToLatitude:          39.991,
		EstimatedDistanceM:  12500,
		EstimatedDurationS:  1800,
		EstimatedPriceCents: 29900,
		Status:              orderproto.OrderStatus_ORDER_STATUS_ACCEPTED,
		CreatedAt:           100,
		UpdatedAt:           200,
	}, nil
	return &orderproto.GetOrderResponse{OrderId: req.OrderId, OrderNo: "NO-1001", FromAddress: "起点", ToAddress: "终点", Status: orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT, EstimatedPriceCents: 29900, CreatedAt: 100}, nil
}

func (f *fakeOrderClient) ListOrders(_ context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	f.listOrdersRequest = req
	return &orderproto.ListOrdersResponse{
		List: []*orderproto.OrderSummary{{
			OrderId:             1001,
			OrderNo:             "NO-1001",
			FromAddress:         "起点",
			ToAddress:           "终点",
			Status:              req.Status,
			EstimatedPriceCents: 29900,
			CreatedAt:           100,
		}},
		Total:    1,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (f *fakeOrderClient) AcceptOrder(_ context.Context, req *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	f.acceptRequest = req
	return &orderproto.AcceptOrderResponse{OrderId: req.OrderId, Status: orderproto.OrderStatus_ORDER_STATUS_ACCEPTED}, nil
}

func (f *fakeOrderClient) CancelOrder(_ context.Context, req *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	f.cancelRequest = req
	return &orderproto.CancelOrderResponse{OrderId: req.OrderId, Status: orderproto.OrderStatus_ORDER_STATUS_CANCELLED}, nil
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
	amount := req.ActualPriceCents
	if f.finishResponseAmountCents > 0 {
		amount = f.finishResponseAmountCents
	}
	return &orderproto.FinishTripResponse{
		OrderId:            req.OrderId,
		Status:             orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY,
		PayableAmountCents: amount,
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

func TestCancelOrderForwardsDriverOperatorAndReason(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.CancelOrder(25, &types.CancelOrderRequest{OrderID: 1001, Reason: "司机临时有事"})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_CANCELLED) {
		t.Fatalf("CancelOrder() response = %+v", resp)
	}
	if client.cancelRequest.GetOperatorId() != 25 ||
		client.cancelRequest.GetOperatorType() != constants.OperatorDriver ||
		client.cancelRequest.GetOrderId() != 1001 ||
		client.cancelRequest.GetReason() != "司机临时有事" {
		t.Fatalf("CancelOrder() request = %+v", client.cancelRequest)
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
	driverClient := &fakeDriverClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client, DriverClient: driverClient})

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
	if len(driverClient.serviceStatusRequests) != 1 ||
		driverClient.serviceStatusRequests[0].GetDriverId() != 25 ||
		driverClient.serviceStatusRequests[0].GetOnlineStatus() != 2 {
		t.Fatalf("SetDriverServiceStatus() requests = %+v", driverClient.serviceStatusRequests)
	}
}

func TestFinishTripForwardsTripMetrics(t *testing.T) {
	client := &fakeOrderClient{finishResponseAmountCents: 3200}
	driverClient := &fakeDriverClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client, DriverClient: driverClient})
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
		client.finishRequest.GetActualPriceCents() != 0 {
		t.Fatalf("FinishTrip() request = %+v", client.finishRequest)
	}
	if len(driverClient.serviceStatusRequests) != 1 ||
		driverClient.serviceStatusRequests[0].GetDriverId() != 25 ||
		driverClient.serviceStatusRequests[0].GetOnlineStatus() != 1 {
		t.Fatalf("SetDriverServiceStatus() requests = %+v", driverClient.serviceStatusRequests)
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
	if _, err := logic.FinishTrip(25, &types.FinishTripRequest{OrderID: 1001, ActualDistanceM: 1, ActualDurationS: 1, ActualPriceCents: -1}); err != nil {
		t.Fatalf("FinishTrip(negative reported price) error = %v, want success", err)
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

func TestListMyOrdersUsesOrderService(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.ListMyOrders(25, 2, 8, int32(orderproto.OrderStatus_ORDER_STATUS_COMPLETED))
	if err != nil {
		t.Fatalf("ListMyOrders() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].OrderNo != "NO-1001" {
		t.Fatalf("ListMyOrders() response = %+v", resp)
	}
	if client.listOrdersRequest.GetDriverId() != 25 ||
		client.listOrdersRequest.GetPage() != 2 ||
		client.listOrdersRequest.GetPageSize() != 8 ||
		client.listOrdersRequest.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_COMPLETED {
		t.Fatalf("ListMyOrders() request = %+v", client.listOrdersRequest)
	}
}

func TestListAvailableOrdersUsesWaitAcceptWithoutDriverFilter(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.ListAvailableOrders(1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].OrderNo != "NO-1001" {
		t.Fatalf("ListAvailableOrders() response = %+v", resp)
	}
	if client.listOrdersRequest.GetDriverId() != 0 ||
		client.listOrdersRequest.GetPage() != 1 ||
		client.listOrdersRequest.GetPageSize() != 10 ||
		client.listOrdersRequest.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT {
		t.Fatalf("ListAvailableOrders() request = %+v", client.listOrdersRequest)
	}
}

func TestGetMyOrderDetailRequiresOrderOwnedByDriver(t *testing.T) {
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetMyOrderDetail(25, 1001)
	if err != nil {
		t.Fatalf("GetMyOrderDetail() error = %v", err)
	}
	if client.getOrderRequest.GetOrderId() != 1001 {
		t.Fatalf("GetMyOrderDetail() request = %+v", client.getOrderRequest)
	}
	if resp.Order.OrderID != 1001 ||
		resp.Order.DriverID != 25 ||
		resp.Order.UserID != 300 ||
		resp.Order.FromLongitude != 116.391 ||
		resp.Order.ToLatitude != 39.991 ||
		resp.Order.EstimatedDistanceM != 12500 ||
		resp.Order.Status != int32(orderproto.OrderStatus_ORDER_STATUS_ACCEPTED) {
		t.Fatalf("GetMyOrderDetail() response = %+v", resp)
	}
}

func TestGetMyOrderDetailRejectsOtherDriverOrder(t *testing.T) {
	client := &fakeOrderClient{getOrderResponseDriverID: 26}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	_, err := logic.GetMyOrderDetail(25, 1001)
	if err != ErrForbiddenDriverResource {
		t.Fatalf("GetMyOrderDetail() error = %v, want %v", err, ErrForbiddenDriverResource)
	}
}
