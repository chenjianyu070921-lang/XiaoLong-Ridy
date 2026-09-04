package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/constants"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	priceproto "XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type fakeOrderClient struct {
	getOrderRequest           *orderproto.GetOrderRequest
	getOrderResponseDriverID  int64
	getOrderResponseStatus    orderproto.OrderStatus
	acceptRequest             *orderproto.AcceptOrderRequest
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
			FromLongitude:       116.391,
			FromLatitude:        39.907,
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

type fakePriceClient struct {
	estimateRequest  *priceproto.EstimatePriceRequest
	estimateResponse *priceproto.EstimatePriceResponse
	estimateErr      error
}

func (f *fakePriceClient) EstimatePrice(_ context.Context, req *priceproto.EstimatePriceRequest) (*priceproto.EstimatePriceResponse, error) {
	f.estimateRequest = req
	if f.estimateErr != nil {
		return nil, f.estimateErr
	}
	if f.estimateResponse != nil {
		return f.estimateResponse, nil
	}
	return &priceproto.EstimatePriceResponse{
		PriceRuleId: 77,
		TotalCents:  4860,
		Detail: &priceproto.PriceDetail{
			BaseFeeCents:     1300,
			DistanceFeeCents: 2600,
			TimeFeeCents:     720,
			NightFeeCents:    0,
			DynamicFeeCents:  240,
			TotalCents:       4860,
		},
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

type fakeOrderTrajectoryRepository struct {
	listDriverID int64
	listOrderID  int64
	records      []svc.TrajectoryRecord
	err          error
}

func (f *fakeOrderTrajectoryRepository) ListByOrder(_ context.Context, driverID, orderID int64) ([]svc.TrajectoryRecord, error) {
	f.listDriverID = driverID
	f.listOrderID = orderID
	if f.err != nil {
		return nil, f.err
	}
	return f.records, nil
}

func (f *fakeOrderTrajectoryRepository) RecordPoint(_ context.Context, _ *svc.TrajectoryRecord) error {
	return nil
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

func TestGetRealtimeFareRequiresOnTripOrderOwnedByDriverAndReturnsPriceDetail(t *testing.T) {
	orderClient := &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_ON_TRIP}
	priceClient := &fakePriceClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{
		OrderClient: orderClient,
		PriceClient: priceClient,
	})

	resp, err := logic.GetRealtimeFare(25, &types.RealtimeFareRequest{
		OrderID:         1001,
		ActualDistanceM: 13800,
		ActualDurationS: 2040,
	})
	if err != nil {
		t.Fatalf("GetRealtimeFare() error = %v", err)
	}
	if resp.OrderID != 1001 || resp.TotalCents != 4860 || resp.PriceRuleID != 77 || resp.Source != "pricesvc" {
		t.Fatalf("GetRealtimeFare() response = %+v", resp)
	}
	if resp.Detail.BaseFeeCents != 1300 || resp.Detail.DynamicFeeCents != 240 || resp.Detail.TotalCents != 4860 {
		t.Fatalf("GetRealtimeFare() detail = %+v", resp.Detail)
	}
	if orderClient.getOrderRequest == nil || orderClient.getOrderRequest.GetOrderId() != 1001 {
		t.Fatalf("GetRealtimeFare() order request = %+v", orderClient.getOrderRequest)
	}
	if priceClient.estimateRequest.GetUserId() != 300 ||
		priceClient.estimateRequest.GetCityCode() != "110000" ||
		priceClient.estimateRequest.GetCarType() != 1 ||
		priceClient.estimateRequest.GetDistanceM() != 13800 ||
		priceClient.estimateRequest.GetDurationS() != 2040 {
		t.Fatalf("GetRealtimeFare() price request = %+v", priceClient.estimateRequest)
	}
}

func TestGetRealtimeFareRejectsOtherDriverOrder(t *testing.T) {
	priceClient := &fakePriceClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{
		OrderClient: &fakeOrderClient{
			getOrderResponseDriverID: 26,
			getOrderResponseStatus:   orderproto.OrderStatus_ORDER_STATUS_ON_TRIP,
		},
		PriceClient: priceClient,
	})

	_, err := logic.GetRealtimeFare(25, &types.RealtimeFareRequest{OrderID: 1001, ActualDistanceM: 1})
	if err != ErrForbiddenDriverResource {
		t.Fatalf("GetRealtimeFare() error = %v, want %v", err, ErrForbiddenDriverResource)
	}
	if priceClient.estimateRequest != nil {
		t.Fatalf("GetRealtimeFare() should not call pricesvc for another driver's order, got %+v", priceClient.estimateRequest)
	}
}

func TestGetRealtimeFareRejectsNonTripOrder(t *testing.T) {
	priceClient := &fakePriceClient{}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{
		OrderClient: &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_ACCEPTED},
		PriceClient: priceClient,
	})

	_, err := logic.GetRealtimeFare(25, &types.RealtimeFareRequest{OrderID: 1001, ActualDistanceM: 1})
	if err != ErrInvalidParam {
		t.Fatalf("GetRealtimeFare() error = %v, want %v", err, ErrInvalidParam)
	}
	if priceClient.estimateRequest != nil {
		t.Fatalf("GetRealtimeFare() should not call pricesvc for non-trip order, got %+v", priceClient.estimateRequest)
	}
}

func TestGetOrderTrajectoryReturnsCurrentDriverPoints(t *testing.T) {
	repo := &fakeOrderTrajectoryRepository{records: []svc.TrajectoryRecord{{
		OrderID:    1001,
		DriverID:   25,
		Longitude:  116.391,
		Latitude:   39.907,
		SpeedKmh:   31.5,
		Heading:    180,
		RecordedAt: time.Unix(123, 0),
	}}}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{TrajectoryRepository: repo})

	resp, err := logic.GetOrderTrajectory(25, 1001)
	if err != nil {
		t.Fatalf("GetOrderTrajectory() error = %v", err)
	}
	if repo.listDriverID != 25 || repo.listOrderID != 1001 {
		t.Fatalf("GetOrderTrajectory() repository query = driver:%d order:%d", repo.listDriverID, repo.listOrderID)
	}
	if resp.OrderID != 1001 || resp.Total != 1 || len(resp.Points) != 1 {
		t.Fatalf("GetOrderTrajectory() response = %+v", resp)
	}
	point := resp.Points[0]
	if point.OrderID != 1001 || point.DriverID != 25 || point.Longitude != 116.391 ||
		point.Latitude != 39.907 || point.SpeedKmh != 31.5 || point.Heading != 180 ||
		point.ReportTime != 123 {
		t.Fatalf("GetOrderTrajectory() point = %+v", point)
	}
}

func TestGetOrderTrajectoryRequiresRepository(t *testing.T) {
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{})

	_, err := logic.GetOrderTrajectory(25, 1001)
	if err != ErrTrajectoryRepositoryNotConfigured {
		t.Fatalf("GetOrderTrajectory() error = %v, want %v", err, ErrTrajectoryRepositoryNotConfigured)
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
	if _, err := logic.GetRealtimeFare(25, &types.RealtimeFareRequest{OrderID: 1001, ActualDurationS: -1}); err != ErrInvalidParam {
		t.Fatalf("GetRealtimeFare(negative duration) error = %v, want %v", err, ErrInvalidParam)
	}
	if _, err := logic.GetOrderTrajectory(25, 0); err != ErrInvalidParam {
		t.Fatalf("GetOrderTrajectory(invalid order) error = %v, want %v", err, ErrInvalidParam)
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
	listResponse  *dispatchproto.ListDispatchRecordsResponse
	listErr       error
}

func (f *fakeDispatchClient) RejectDispatch(_ context.Context, req *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error) {
	f.rejectRequest = req
	return &dispatchproto.RejectDispatchResponse{OrderId: req.OrderId, DriverId: req.DriverId, Status: 2}, nil
}

func (f *fakeDispatchClient) ListDispatchRecords(_ context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	f.listRequest = req
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listResponse != nil {
		return f.listResponse, nil
	}
	return &dispatchproto.ListDispatchRecordsResponse{List: []*dispatchproto.DispatchRecord{{Id: 7, OrderId: 1001, DriverId: req.DriverId, Status: 1, Remark: "dispatch note", RejectReason: "too far"}}, Total: 1, Page: req.Page, PageSize: req.PageSize}, nil
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

// TestListAvailableOrdersReadsAvailableSetAndGetOrder verifies nearby orders come from
// driver:available:%d + GetOrder instead of global ListOrders.
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

func TestListAvailableOrdersReturnsEmptyWhenAvailableSetEmpty(t *testing.T) {
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
	client := &fakeOrderClient{}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if client.listOrdersRequest != nil {
		t.Fatalf("ListAvailableOrders() should not call global ListOrders when available set is empty, got %+v", client.listOrdersRequest)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListAvailableOrders() response = %+v, want empty list when driver has no assigned orders", resp)
	}
}

func TestListAvailableOrdersFallsBackToPendingDispatchRecordsWhenAvailableSetEmpty(t *testing.T) {
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
	orderClient := &fakeOrderClient{
		getOrderResponses: map[int64]*orderproto.GetOrderResponse{
			1001: availableOrderDetail(1001, "pending-dispatch-fallback", 116.398, 39.908),
		},
	}
	dispatchClient := &fakeDispatchClient{listResponse: &dispatchproto.ListDispatchRecordsResponse{
		List: []*dispatchproto.DispatchRecord{{
			Id:       7,
			OrderId:  1001,
			DriverId: driverID,
			Status:   constants.DispatchStatusPending,
		}},
		Total:    1,
		Page:     1,
		PageSize: 1000,
	}}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: orderClient, DispatchClient: dispatchClient, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].OrderID != 1001 {
		t.Fatalf("ListAvailableOrders() response = %+v, want pending dispatch order 1001", resp)
	}
	if dispatchClient.listRequest == nil || dispatchClient.listRequest.GetStatus() != constants.DispatchStatusPending {
		t.Fatalf("ListAvailableOrders() dispatch query = %+v, want pending dispatch query", dispatchClient.listRequest)
	}
}

func TestListAvailableOrdersReturnsEmptyWhenDriverPositionInvalid(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	driverID := int64(25)
	if err := rdb.SAdd(ctx, constants.RedisDriverOnline, fmt.Sprint(driverID)).Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	if err := rdb.HSet(ctx, fmt.Sprintf(constants.RedisDriverPos, driverID), map[string]interface{}{
		"longitude": "0",
		"latitude":  "0",
	}).Err(); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}
	if err := rdb.SAdd(ctx, fmt.Sprintf(constants.RedisDriverAvailable, driverID), "1001").Err(); err != nil {
		t.Fatalf("SAdd available() error = %v", err)
	}
	client := &fakeOrderClient{
		getOrderResponses: map[int64]*orderproto.GetOrderResponse{
			1001: availableOrderDetail(1001, "zero", 0, 0),
		},
	}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListAvailableOrders() response = %+v, want empty for invalid position", resp)
	}
}

func TestListAvailableOrdersHidesOrdersWithoutPendingDispatch(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	driverID := int64(25)
	availableKey := fmt.Sprintf(constants.RedisDriverAvailable, driverID)
	if err := rdb.SAdd(ctx, constants.RedisDriverOnline, fmt.Sprint(driverID)).Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	if err := rdb.HSet(ctx, fmt.Sprintf(constants.RedisDriverPos, driverID), map[string]interface{}{
		"longitude": "116.397",
		"latitude":  "39.908",
	}).Err(); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}
	if err := rdb.SAdd(ctx, availableKey, "1001").Err(); err != nil {
		t.Fatalf("SAdd available() error = %v", err)
	}
	client := &fakeOrderClient{
		getOrderResponses: map[int64]*orderproto.GetOrderResponse{
			1001: availableOrderDetail(1001, "wait-without-pending-dispatch", 116.398, 39.908),
		},
	}
	dispatchClient := &fakeDispatchClient{listResponse: &dispatchproto.ListDispatchRecordsResponse{
		List: []*dispatchproto.DispatchRecord{},
	}}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, DispatchClient: dispatchClient, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListAvailableOrders() response = %+v, want empty without pending dispatch", resp)
	}
	if dispatchClient.listRequest == nil ||
		dispatchClient.listRequest.GetDriverId() != driverID ||
		dispatchClient.listRequest.GetStatus() != constants.DispatchStatusPending {
		t.Fatalf("ListAvailableOrders() dispatch query = %+v, want current driver's pending dispatch records", dispatchClient.listRequest)
	}
	if ok, err := rdb.SIsMember(ctx, availableKey, "1001").Result(); err != nil || ok {
		t.Fatalf("stale available order should be removed from Redis set, exists=%v err=%v", ok, err)
	}
}

// TestListAvailableOrdersHidesAcceptedOrders verifies accepted orders are filtered out.
func TestListAvailableOrdersHidesAcceptedOrders(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	ctx := context.Background()
	driverID := int64(25)
	availableKey := fmt.Sprintf(constants.RedisDriverAvailable, driverID)
	if err := rdb.SAdd(ctx, constants.RedisDriverOnline, fmt.Sprint(driverID)).Err(); err != nil {
		t.Fatalf("SAdd() error = %v", err)
	}
	if err := rdb.HSet(ctx, fmt.Sprintf(constants.RedisDriverPos, driverID), map[string]interface{}{
		"longitude": "116.397",
		"latitude":  "39.908",
	}).Err(); err != nil {
		t.Fatalf("HSet() error = %v", err)
	}
	if err := rdb.SAdd(ctx, availableKey, "1001").Err(); err != nil {
		t.Fatalf("SAdd available() error = %v", err)
	}
	client := &fakeOrderClient{getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_ACCEPTED}
	dispatchClient := &fakeDispatchClient{listResponse: &dispatchproto.ListDispatchRecordsResponse{
		List: []*dispatchproto.DispatchRecord{{
			Id:       7,
			OrderId:  1001,
			DriverId: driverID,
			Status:   constants.DispatchStatusPending,
		}},
	}}
	logic := NewOrderLogic(ctx, &svc.ServiceContext{OrderClient: client, DispatchClient: dispatchClient, RedisClient: rdb})

	resp, err := logic.ListAvailableOrders(driverID, 1, 10)
	if err != nil {
		t.Fatalf("ListAvailableOrders() error = %v", err)
	}
	if resp.Total != 0 || len(resp.List) != 0 {
		t.Fatalf("ListAvailableOrders() response = %+v, want empty when order is already accepted", resp)
	}
	if dispatchClient.listRequest == nil {
		t.Fatal("ListAvailableOrders() should query pending dispatches before filtering")
	}
	if ok, err := rdb.SIsMember(ctx, availableKey, "1001").Result(); err != nil || ok {
		t.Fatalf("accepted order should be removed from Redis set, exists=%v err=%v", ok, err)
	}
}

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
	// available set seed; 1002 is filtered out because it is too far away.
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
	// New logic does not call ListOrders.
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
		resp.Order.UserID != 0 || // UserID is hidden from drivers and should be 0.
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
