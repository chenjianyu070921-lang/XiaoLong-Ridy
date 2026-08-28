package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/constants"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type fakeOrderClient struct {
	getOrderRequest           *orderproto.GetOrderRequest
	getOrderResponseDriverID  int64
	getOrderResponseStatus    orderproto.OrderStatus
	acceptRequest             *orderproto.AcceptOrderRequest
	cancelRequest             *orderproto.CancelOrderRequest
	arriveRequest             *orderproto.ConfirmArriveRequest
	startRequest              *orderproto.StartTripRequest
	finishRequest             *orderproto.FinishTripRequest
	finishResponseAmountCents int64
	listOrdersRequest         *orderproto.ListOrdersRequest
	listOrdersResponse        *orderproto.ListOrdersResponse
	getOrderResponses         map[int64]*orderproto.GetOrderResponse
	getOrderErr               error
}

func (f *fakeOrderClient) GetOrder(_ context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	f.getOrderRequest = req
	if f.getOrderErr != nil {
		return nil, f.getOrderErr
	}
	if f.getOrderResponses != nil {
		if resp, ok := f.getOrderResponses[req.GetOrderId()]; ok {
			return resp, nil
		}
	}
	driverID := f.getOrderResponseDriverID
	if driverID == 0 {
		driverID = 25
	}
	if driverID < 0 {
		driverID = 0
	}
	status := f.getOrderResponseStatus
	if status == orderproto.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		status = orderproto.OrderStatus_ORDER_STATUS_ACCEPTED
	}
	return &orderproto.GetOrderResponse{
		OrderId:             req.OrderId,
		OrderNo:             "NO-1001",
		UserId:              300,
		DriverId:            driverID,
		CarType:             1,
		FromAddress:         "pickup",
		FromLongitude:       116.391,
		FromLatitude:        39.907,
		ToAddress:           "destination",
		ToLongitude:         116.481,
		ToLatitude:          39.991,
		EstimatedDistanceM:  12500,
		EstimatedDurationS:  1800,
		EstimatedPriceCents: 29900,
		PayableCents:        1001,
		PaidCents:           1001,
		Status:              status,
		CreatedAt:           100,
		UpdatedAt:           200,
	}, nil
}

func (f *fakeOrderClient) ListOrders(_ context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	f.listOrdersRequest = req
	if f.listOrdersResponse != nil {
		return f.listOrdersResponse, nil
	}
	return &orderproto.ListOrdersResponse{
		List: []*orderproto.OrderSummary{{
			OrderId:             1001,
			OrderNo:             "NO-1001",
			FromAddress:         "pickup",
			ToAddress:           "destination",
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

	resp, err := logic.CancelOrder(25, &types.CancelOrderRequest{OrderID: 1001, Reason: "driver unavailable"})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_CANCELLED) {
		t.Fatalf("CancelOrder() response = %+v", resp)
	}
	if client.cancelRequest.GetOperatorId() != 25 ||
		client.cancelRequest.GetOperatorType() != constants.OperatorDriver ||
		client.cancelRequest.GetOrderId() != 1001 ||
		client.cancelRequest.GetReason() != "driver unavailable" {
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
		OrderID:         1001,
		ActualDistanceM: 12500,
		ActualDurationS: 1800,
	}

	resp, err := logic.FinishTrip(25, req)
	if err != nil {
		t.Fatalf("FinishTrip() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.Status != int32(orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY) || resp.PayableAmountCents != 3200 {
		t.Fatalf("FinishTrip() response = %+v", resp)
	}
	if client.finishRequest.GetDriverId() != 25 || client.finishRequest.GetOrderId() != 1001 ||
		client.finishRequest.GetActualDistanceM() != 12500 || client.finishRequest.GetActualDurationS() != 1800 {
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
}

func TestRejectOrderRequiresDispatchClient(t *testing.T) {
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: &fakeDispatchClient{}})

	_, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "too far"})
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
	return &dispatchproto.ListDispatchRecordsResponse{List: []*dispatchproto.DispatchRecord{{Id: 7, OrderId: 1001, DriverId: req.DriverId, Status: 1, Remark: "too far"}}, Total: 1, Page: req.Page, PageSize: req.PageSize}, nil
}

func TestRejectOrderForwardsDispatchRequest(t *testing.T) {
	client := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

	resp, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "temporary conflict"})
	if err != nil || resp.OrderID != 1001 || resp.DriverID != 25 {
		t.Fatalf("RejectOrder() response = %+v, error = %v", resp, err)
	}
	if client.rejectRequest.GetOrderId() != 1001 || client.rejectRequest.GetDriverId() != 25 || client.rejectRequest.GetReason() != "temporary conflict" {
		t.Fatalf("RejectOrder() request = %+v", client.rejectRequest)
	}
}

func TestRejectOrderRequiresReason(t *testing.T) {
	client := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

	_, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "   "})
	if err != ErrInvalidParam {
		t.Fatalf("RejectOrder() error = %v, want %v", err, ErrInvalidParam)
	}
	if client.rejectRequest != nil {
		t.Fatalf("RejectOrder() should not call dispatchsvc for empty reason, got %+v", client.rejectRequest)
	}
}

func TestRejectOrderTrimsReasonBeforeDispatch(t *testing.T) {
	client := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

	_, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "  too far  "})
	if err != nil {
		t.Fatalf("RejectOrder() error = %v", err)
	}
	if client.rejectRequest.GetReason() != "too far" {
		t.Fatalf("RejectOrder() reason = %q, want too far", client.rejectRequest.GetReason())
	}
}

func TestListMyDispatchesCombinesDispatchAndOrder(t *testing.T) {
	dispatchClient := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: dispatchClient, OrderClient: &fakeOrderClient{}})

	resp, err := logic.ListMyDispatches(25, 1, 20, 1)
	if err != nil || len(resp.List) != 1 || resp.List[0].Order.OrderNo != "NO-1001" {
		t.Fatalf("ListMyDispatches() response = %+v, error = %v", resp, err)
	}

	if resp.List[0].Dispatch.RejectReason != "too far" {
		t.Fatalf("ListMyDispatches() reject reason = %q, want too far", resp.List[0].Dispatch.RejectReason)
	}
	if dispatchClient.listRequest.GetDriverId() != 25 || dispatchClient.listRequest.GetStatus() != 1 {
		t.Fatalf("ListMyDispatches() request = %+v", dispatchClient.listRequest)
	}
	if !resp.OrderQueryOk {
		t.Fatalf("ListMyDispatches() orderQueryOk = false, want true")
	}
}

func TestListMyDispatchesMarksOrderQueryFailed(t *testing.T) {
	dispatchClient := &fakeDispatchClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{
		DispatchClient: dispatchClient,
		OrderClient:    &fakeOrderClient{getOrderErr: errors.New("ordersvc unavailable")},
	})

	resp, err := logic.ListMyDispatches(25, 1, 20, 1)
	if err != nil {
		t.Fatalf("ListMyDispatches() error = %v", err)
	}
	if resp.OrderQueryOk {
		t.Fatalf("ListMyDispatches() orderQueryOk = true, want false")
	}
	if len(resp.List) != 1 || resp.List[0].Dispatch.OrderID != 1001 {
		t.Fatalf("ListMyDispatches() should keep dispatch record when order query fails: %+v", resp)
	}
	if resp.List[0].Order.OrderID != 0 || resp.List[0].Order.OrderNo != "" {
		t.Fatalf("ListMyDispatches() order should be zero value on query failure: %+v", resp.List[0].Order)
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

// TestListAvailableOrdersReadsAvailableSetAndGetOrder 验证 D3/D8 修复后：
// ListAvailableOrders 读 driver:available:%d 集合 + GetOrder，而非全局 ListOrders。
func TestListAvailableOrdersReadsAvailableSetAndGetOrder(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	if err := rdb.SAdd(ctx, constants.RedisDriverOnline, "25").Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	if err := rdb.HSet(ctx, fmt.Sprintf(constants.RedisDriverPos, 25), map[string]interface{}{
		"longitude": "116.397",
		"latitude":  "39.908",
	}).Err(); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}
	// 派单消费者写入的 available 集合
	if err := rdb.SAdd(ctx, fmt.Sprintf(constants.RedisDriverAvailable, 25), "1001").Err(); err != nil {
		t.Fatalf("SAdd available() error = %v", err)
	}
	client := &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(25, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].OrderNo != "NO-1001" {
		t.Fatalf("ListAvailableOrders() response = %+v", resp)
	}
	// 新逻辑不调用 ListOrders，而是读 available 集合 + GetOrder
	if client.listOrdersRequest != nil {
		t.Fatalf("ListAvailableOrders() should not call ListOrders, got %+v", client.listOrdersRequest)
	}
	if client.getOrderRequest == nil || client.getOrderRequest.GetOrderId() != 1001 {
		t.Fatalf("ListAvailableOrders() should GetOrder(1001), got %+v", client.getOrderRequest)
	}
}

func TestListAvailableOrdersReturnsEmptyWhenDriverOffline(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client := &fakeOrderClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(25, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListAvailableOrders() response = %+v, want empty", resp)
	}
	if client.listOrdersRequest != nil {
		t.Fatalf("ListAvailableOrders() should not call ordersvc for offline driver, got %+v", client.listOrdersRequest)
	}
}

// TestListAvailableOrdersFiltersAndSortsByDriverPosition 验证距离过滤与排序仍然生效（新逻辑保留距离计算）。
func TestListAvailableOrdersFiltersAndSortsByDriverPosition(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	driverID := int64(25)
	if err := rdb.SAdd(ctx, constants.RedisDriverOnline, fmt.Sprint(driverID)).Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	if err := rdb.HSet(ctx, fmt.Sprintf(constants.RedisDriverPos, driverID), map[string]interface{}{
		"longitude": "116.397",
		"latitude":  "39.908",
	}).Err(); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}
	// 派单消费者写入的 available 集合（3 个订单，其中 1002 距离超过 3km 应被过滤）
	if err := rdb.SAdd(ctx, fmt.Sprintf(constants.RedisDriverAvailable, driverID), "1001", "1002", "1003").Err(); err != nil {
		t.Fatalf("SAdd available() error = %v", err)
	}
	client := &fakeOrderClient{
		getOrderResponses: map[int64]*orderproto.GetOrderResponse{
			1001: availableOrderDetail(1001, "far-but-in-range", 116.415, 39.908),
			1002: availableOrderDetail(1002, "too-far", 116.520, 39.908),
			1003: availableOrderDetail(1003, "nearest", 116.398, 39.908),
		},
	}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 2 || len(resp.List) != 2 {
		t.Fatalf("ListAvailableOrders() response = %+v, want 2 nearby orders", resp)
	}
	if got := []int64{resp.List[0].OrderID, resp.List[1].OrderID}; got[0] != 1003 || got[1] != 1001 {
		t.Fatalf("ListAvailableOrders() order IDs = %v, want [1003 1001]", got)
	}
	if resp.List[0].DistanceMeters <= 0 || resp.List[0].DistanceMeters >= resp.List[1].DistanceMeters {
		t.Fatalf("ListAvailableOrders() distances = %d/%d, want nearest first", resp.List[0].DistanceMeters, resp.List[1].DistanceMeters)
	}
	// 新逻辑不调用 ListOrders
	if client.listOrdersRequest != nil {
		t.Fatalf("ListAvailableOrders() should not call ListOrders, got %+v", client.listOrdersRequest)
	}
}

func availableOrderDetail(orderID int64, orderNo string, longitude, latitude float64) *orderproto.GetOrderResponse {
	return &orderproto.GetOrderResponse{
		OrderId:             orderID,
		OrderNo:             orderNo,
		UserId:              300,
		CarType:             1,
		FromAddress:         "pickup",
		FromLongitude:       longitude,
		FromLatitude:        latitude,
		ToAddress:           "destination",
		ToLongitude:         116.481,
		ToLatitude:          39.991,
		EstimatedDistanceM:  12500,
		EstimatedDurationS:  1800,
		EstimatedPriceCents: 29900,
		Status:              orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		CreatedAt:           100,
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
		resp.Order.UserID != 0 || // UserID 对司机隐藏，应为 0
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

func TestGetMyOrderDetailAllowsUnassignedWaitAcceptOrder(t *testing.T) {
	client := &fakeOrderClient{
		getOrderResponseDriverID: -1,
		getOrderResponseStatus:   orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
	}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetMyOrderDetail(25, 1001)
	if err != nil {
		t.Fatalf("GetMyOrderDetail() error = %v", err)
	}
	if client.getOrderRequest.GetOrderId() != 1001 {
		t.Fatalf("GetMyOrderDetail() request = %+v", client.getOrderRequest)
	}
	if resp.Order.DriverID != 0 || resp.Order.Status != int32(orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT) {
		t.Fatalf("GetMyOrderDetail() response = %+v", resp)
	}
}
