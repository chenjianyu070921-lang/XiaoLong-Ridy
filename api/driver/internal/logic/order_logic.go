package logic

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/constants"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	priceproto "XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *OrderLogic) AcceptOrder(driverID, orderID int64) (*types.AcceptOrderResponse, error) {
	if driverID <= 0 || orderID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.AcceptOrder(l.ctx, &orderproto.AcceptOrderRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	// 接单后从 available 集合移除，避免大厅/WS 重复展示已接订单
	l.removeFromAvailable(driverID, orderID)
	return &types.AcceptOrderResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

func (l *OrderLogic) StartTrip(driverID, orderID int64) (*types.StartTripResponse, error) {
	if driverID <= 0 || orderID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.StartTrip(l.ctx, &orderproto.StartTripRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	if driverClient, err := l.driverClient(); err == nil {
		// 订单已成功开始，司机状态同步为尽力而为（best-effort）：同步失败不应让已成功的
		// 订单返回错误。重试一次后仍失败则记录结构化告警（含 orderId/driverId）以便对账。
		if _, sErr := driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
			DriverId:     driverID,
			OnlineStatus: 2,
		}); sErr != nil {
			if _, retryErr := driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
				DriverId:     driverID,
				OnlineStatus: 2,
			}); retryErr != nil {
				logx.WithContext(l.ctx).Errorf("set driver on-trip after start failed after retry (orderId=%d, driverId=%d): %v", orderID, driverID, retryErr)
			}
		}
	}
	return &types.StartTripResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

func (l *OrderLogic) ConfirmArrive(driverID, orderID int64) (*types.ConfirmArriveResponse, error) {
	if driverID <= 0 || orderID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ConfirmArrive(l.ctx, &orderproto.ConfirmArriveRequest{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		return nil, err
	}
	return &types.ConfirmArriveResponse{
		OrderID: resp.GetOrderId(),
		Status:  int32(resp.GetStatus()),
	}, nil
}

func (l *OrderLogic) FinishTrip(driverID int64, req *types.FinishTripRequest) (*types.FinishTripResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 || req.ActualDistanceM < 0 || req.ActualDurationS < 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.FinishTrip(l.ctx, &orderproto.FinishTripRequest{
		OrderId:         req.OrderID,
		DriverId:        driverID,
		ActualDistanceM: req.ActualDistanceM,
		ActualDurationS: req.ActualDurationS,
	})
	if err != nil {
		return nil, err
	}
	if driverClient, err := l.driverClient(); err == nil {
		// 订单已成功结束，司机状态同步为尽力而为（best-effort）：同步失败不应让已成功的
		// 订单返回错误。重试一次后仍失败则记录结构化告警（含 orderId/driverId）以便对账。
		if _, sErr := driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
			DriverId:     driverID,
			OnlineStatus: 1,
		}); sErr != nil {
			if _, retryErr := driverClient.SetDriverServiceStatus(l.ctx, &driversproto.SetDriverServiceStatusRequest{
				DriverId:     driverID,
				OnlineStatus: 1,
			}); retryErr != nil {
				logx.WithContext(l.ctx).Errorf("set driver online after finish failed after retry (orderId=%d, driverId=%d): %v", req.OrderID, driverID, retryErr)
			}
		}
	}
	return &types.FinishTripResponse{
		OrderID:            resp.GetOrderId(),
		Status:             int32(resp.GetStatus()),
		PayableAmountCents: resp.GetPayableAmountCents(),
	}, nil
}

func (l *OrderLogic) GetRealtimeFare(driverID int64, req *types.RealtimeFareRequest) (*types.RealtimeFareResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 || req.ActualDistanceM < 0 || req.ActualDurationS < 0 {
		return nil, ErrInvalidParam
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	priceClient, err := l.priceClient()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetDriverId() != driverID {
		return nil, ErrForbiddenDriverResource
	}
	if order.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_ON_TRIP {
		return nil, ErrInvalidParam
	}
	resp, err := priceClient.EstimatePrice(l.ctx, &priceproto.EstimatePriceRequest{
		UserId:    order.GetUserId(),
		CityCode:  order.GetCityCode(),
		CarType:   order.GetCarType(),
		DistanceM: req.ActualDistanceM,
		DurationS: req.ActualDurationS,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, ErrInvalidParam
	}
	detail := resp.GetDetail()
	mapped := types.RealtimeFareDetail{}
	if detail != nil {
		mapped = types.RealtimeFareDetail{
			BaseFeeCents:     detail.GetBaseFeeCents(),
			DistanceFeeCents: detail.GetDistanceFeeCents(),
			TimeFeeCents:     detail.GetTimeFeeCents(),
			NightFeeCents:    detail.GetNightFeeCents(),
			DynamicFeeCents:  detail.GetDynamicFeeCents(),
			TotalCents:       detail.GetTotalCents(),
		}
	}
	totalCents := resp.GetTotalCents()
	if totalCents == 0 {
		totalCents = mapped.TotalCents
	}
	return &types.RealtimeFareResponse{
		OrderID:     order.GetOrderId(),
		PriceRuleID: resp.GetPriceRuleId(),
		TotalCents:  totalCents,
		Source:      "pricesvc",
		Detail:      mapped,
	}, nil
}

func (l *OrderLogic) RejectOrder(driverID int64, req *types.RejectOrderRequest) (*types.RejectOrderResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidParam
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, ErrInvalidParam
	}
	client, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.RejectDispatch(l.ctx, &dispatchproto.RejectDispatchRequest{
		OrderId:  req.OrderID,
		DriverId: driverID,
		Reason:   reason,
	})
	if err != nil {
		return nil, err
	}
	// 拒单后从 available 集合移除，避免大厅/WS 在 TTL 过期前重复展示已拒订单
	l.removeFromAvailable(driverID, req.OrderID)
	return &types.RejectOrderResponse{
		OrderID:  resp.GetOrderId(),
		DriverID: resp.GetDriverId(),
		Status:   resp.GetStatus(),
	}, nil
}

func (l *OrderLogic) ListMyDispatches(driverID int64, page, pageSize, status int32) (*types.ListMyDispatchesResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, err
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPage(page, pageSize)
	resp, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		DriverId: driverID,
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.MyDispatchItem, 0, len(resp.GetList()))
	orderQueryOk := true
	for _, record := range resp.GetList() {
		item := types.MyDispatchItem{
			Dispatch: types.DispatchRecord{
				ID:           record.GetId(),
				OrderID:      record.GetOrderId(),
				DriverID:     record.GetDriverId(),
				DispatchType: record.GetDispatchType(),
				Status:       record.GetStatus(),
				MatchScore:   record.GetMatchScore(),
				Remark:       record.GetRemark(),
				RejectReason: record.GetRejectReason(),
				CreatedAt:    record.GetCreatedAt(),
				UpdatedAt:    record.GetUpdatedAt(),
			},
		}
		order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: record.GetOrderId()})
		if err != nil {
			orderQueryOk = false
			logx.WithContext(l.ctx).Errorf("get order for dispatch record failed (orderId=%d): %v", record.GetOrderId(), err)
		} else {
			item.Order = types.OrderBrief{
				OrderID:             order.GetOrderId(),
				OrderNo:             order.GetOrderNo(),
				FromAddress:         order.GetFromAddress(),
				FromLongitude:       order.GetFromLongitude(),
				FromLatitude:        order.GetFromLatitude(),
				ToAddress:           order.GetToAddress(),
				ToLongitude:         order.GetToLongitude(),
				ToLatitude:          order.GetToLatitude(),
				Status:              int32(order.GetStatus()),
				EstimatedPriceCents: order.GetEstimatedPriceCents(),
				CreatedAt:           order.GetCreatedAt(),
			}
		}
		items = append(items, item)
	}
	return &types.ListMyDispatchesResponse{
		List:         items,
		Total:        resp.GetTotal(),
		Page:         resp.GetPage(),
		PageSize:     resp.GetPageSize(),
		OrderQueryOk: orderQueryOk,
	}, nil
}

func (l *OrderLogic) ListMyOrders(driverID int64, page, pageSize, status int32) (*types.ListMyOrdersResponse, error) {
	if driverID <= 0 || status < 0 || status > int32(orderproto.OrderStatus_ORDER_STATUS_CANCELLED) {
		return nil, ErrInvalidParam
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	page, pageSize = clampPage(page, pageSize)
	resp, err := orderClient.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus(status),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.OrderBrief, 0, len(resp.GetList()))
	for _, order := range resp.GetList() {
		items = append(items, types.OrderBrief{
			OrderID:             order.GetOrderId(),
			OrderNo:             order.GetOrderNo(),
			FromAddress:         order.GetFromAddress(),
			FromLongitude:       order.GetFromLongitude(),
			FromLatitude:        order.GetFromLatitude(),
			ToAddress:           order.GetToAddress(),
			ToLongitude:         order.GetToLongitude(),
			ToLatitude:          order.GetToLatitude(),
			Status:              int32(order.GetStatus()),
			EstimatedPriceCents: order.GetEstimatedPriceCents(),
			CreatedAt:           order.GetCreatedAt(),
		})
	}
	return &types.ListMyOrdersResponse{
		List:     items,
		Total:    resp.GetTotal(),
		Page:     resp.GetPage(),
		PageSize: resp.GetPageSize(),
	}, nil
}

// ListAvailableOrders 司机端"可接单列表"：读取派单消费者写入的 driver:available:%d 集合，
// 只展示派给当前司机的待接单订单（与推送链路统一），确保大厅展示的订单司机一定能接单（D3/D8 修复）。
func (l *OrderLogic) ListAvailableOrders(driverID int64, page, pageSize int32) (*types.ListMyOrdersResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	page, pageSize = clampPage(page, pageSize)
	if l.svcCtx == nil || l.svcCtx.RedisClient == nil {
		return emptyOrderList(page, pageSize), nil
	}

	// 校验司机在线
	online, err := l.svcCtx.RedisClient.SIsMember(l.ctx, constants.RedisDriverOnline, fmt.Sprint(driverID)).Result()
	if err != nil || !online {
		return emptyOrderList(page, pageSize), nil
	}

	// 获取司机当前位置（用于距离展示与排序）
	pos, err := l.svcCtx.RedisClient.HGetAll(l.ctx, fmt.Sprintf(constants.RedisDriverPos, driverID)).Result()
	if err != nil {
		return emptyOrderList(page, pageSize), nil
	}
	driverLongitude, driverLatitude, ok := parseDriverPosition(pos)
	if !ok {
		return emptyOrderList(page, pageSize), nil
	}

	// 读取派给当前司机的订单 ID 集合（dispatch_consumer 在派单时 SAdd，90s TTL）
	availableKey := fmt.Sprintf(constants.RedisDriverAvailable, driverID)
	orderIDStrs, err := l.svcCtx.RedisClient.SMembers(l.ctx, availableKey).Result()
	if err != nil {
		return emptyOrderList(page, pageSize), nil
	}

	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	pendingDispatches, err := l.pendingDispatchOrderIDs(driverID)
	if err != nil {
		return nil, err
	}
	if len(orderIDStrs) == 0 {
		orderIDStrs = orderIDStringsFromPendingDispatches(pendingDispatches)
		if len(orderIDStrs) == 0 {
			return emptyOrderList(page, pageSize), nil
		}
	}
	items := l.availableOrdersFromAssignedSet(orderClient, orderIDStrs, pendingDispatches, driverID, driverLongitude, driverLatitude)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].DistanceMeters == items[j].DistanceMeters {
			return items[i].CreatedAt < items[j].CreatedAt
		}
		return items[i].DistanceMeters < items[j].DistanceMeters
	})
	total := int64(len(items))
	items = paginateOrderBriefs(items, page, pageSize)
	return &types.ListMyOrdersResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (l *OrderLogic) pendingDispatchOrderIDs(driverID int64) (map[int64]struct{}, error) {
	dispatchClient, err := l.dispatchClient()
	if err != nil {
		return nil, nil
	}
	resp, err := dispatchClient.ListDispatchRecords(l.ctx, &dispatchproto.ListDispatchRecordsRequest{
		DriverId: driverID,
		Status:   constants.DispatchStatusPending,
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}
	pending := make(map[int64]struct{}, len(resp.GetList()))
	for _, record := range resp.GetList() {
		if record.GetOrderId() > 0 && record.GetStatus() == constants.DispatchStatusPending {
			pending[record.GetOrderId()] = struct{}{}
		}
	}
	return pending, nil
}

func orderIDStringsFromPendingDispatches(pending map[int64]struct{}) []string {
	orderIDStrs := make([]string, 0, len(pending))
	for orderID := range pending {
		if orderID > 0 {
			orderIDStrs = append(orderIDStrs, strconv.FormatInt(orderID, 10))
		}
	}
	sort.Strings(orderIDStrs)
	return orderIDStrs
}

func (l *OrderLogic) availableOrdersFromAssignedSet(orderClient svc.OrderClient, orderIDStrs []string, pendingDispatches map[int64]struct{}, driverID int64, driverLongitude, driverLatitude float64) []types.OrderBrief {
	items := make([]types.OrderBrief, 0, len(orderIDStrs))
	for _, idStr := range orderIDStrs {
		orderID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || orderID <= 0 {
			continue
		}
		if pendingDispatches != nil {
			if _, ok := pendingDispatches[orderID]; !ok {
				l.removeFromAvailable(driverID, orderID)
				continue
			}
		}
		detail, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: orderID})
		if err != nil {
			continue
		}
		// 只展示仍处于待接单状态的订单（已被其他司机接走/已取消的自动过滤）
		if detail.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT {
			l.removeFromAvailable(driverID, orderID)
			continue
		}
		distance := 0.0
		if driverLongitude != 0 || driverLatitude != 0 {
			distance = haversineMeters(driverLongitude, driverLatitude, detail.GetFromLongitude(), detail.GetFromLatitude())
		}
		if distance > availableOrderRadiusMeters {
			continue
		}
		items = append(items, types.OrderBrief{
			OrderID:             detail.GetOrderId(),
			OrderNo:             detail.GetOrderNo(),
			FromAddress:         detail.GetFromAddress(),
			FromLongitude:       detail.GetFromLongitude(),
			FromLatitude:        detail.GetFromLatitude(),
			ToAddress:           detail.GetToAddress(),
			ToLongitude:         detail.GetToLongitude(),
			ToLatitude:          detail.GetToLatitude(),
			Status:              int32(detail.GetStatus()),
			EstimatedPriceCents: detail.GetEstimatedPriceCents(),
			DistanceMeters:      int64(math.Round(distance)),
			CreatedAt:           detail.GetCreatedAt(),
		})
	}
	return items
}

const availableOrderRadiusMeters = 10000.0

func emptyOrderList(page, pageSize int32) *types.ListMyOrdersResponse {
	return &types.ListMyOrdersResponse{
		List:     []types.OrderBrief{},
		Total:    0,
		Page:     page,
		PageSize: pageSize,
	}
}

func parseDriverPosition(pos map[string]string) (float64, float64, bool) {
	longitude, err := strconv.ParseFloat(pos["longitude"], 64)
	if err != nil {
		return 0, 0, false
	}
	latitude, err := strconv.ParseFloat(pos["latitude"], 64)
	if err != nil {
		return 0, 0, false
	}
	if !validLocation(longitude, latitude) {
		return 0, 0, false
	}
	return longitude, latitude, true
}

func haversineMeters(lon1, lat1, lon2, lat2 float64) float64 {
	const earthRadiusMeters = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func paginateOrderBriefs(items []types.OrderBrief, page, pageSize int32) []types.OrderBrief {
	start := int((page - 1) * pageSize)
	if start >= len(items) {
		return []types.OrderBrief{}
	}
	end := start + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func (l *OrderLogic) GetMyOrderDetail(driverID, orderID int64) (*types.GetMyOrderDetailResponse, error) {
	if driverID <= 0 || orderID <= 0 {
		return nil, ErrInvalidParam
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return nil, err
	}
	if !canDriverViewOrder(driverID, order) {
		return nil, ErrForbiddenDriverResource
	}
	return &types.GetMyOrderDetailResponse{Order: types.OrderDetail{
		OrderID:             order.GetOrderId(),
		OrderNo:             order.GetOrderNo(),
		UserID:              0, // 对司机隐藏乘客 UserID，避免隐私泄露
		DriverID:            order.GetDriverId(),
		CarType:             order.GetCarType(),
		FromAddress:         order.GetFromAddress(),
		FromLongitude:       order.GetFromLongitude(),
		FromLatitude:        order.GetFromLatitude(),
		ToAddress:           order.GetToAddress(),
		ToLongitude:         order.GetToLongitude(),
		ToLatitude:          order.GetToLatitude(),
		EstimatedDistanceM:  order.GetEstimatedDistanceM(),
		EstimatedDurationS:  order.GetEstimatedDurationS(),
		EstimatedPriceCents: order.GetEstimatedPriceCents(),
		Status:              int32(order.GetStatus()),
		CancelReason:        order.GetCancelReason(),
		CancelBy:            order.GetCancelBy(),
		CreatedAt:           order.GetCreatedAt(),
		UpdatedAt:           order.GetUpdatedAt(),
	}}, nil
}

func (l *OrderLogic) GetOrderTrajectory(driverID, orderID int64) (*types.GetOrderTrajectoryResponse, error) {
	if driverID <= 0 || orderID <= 0 {
		return nil, ErrInvalidParam
	}
	if l.svcCtx == nil || l.svcCtx.TrajectoryRepository == nil {
		return nil, ErrTrajectoryRepositoryNotConfigured
	}
	records, err := l.svcCtx.TrajectoryRepository.ListByOrder(l.ctx, driverID, orderID)
	if err != nil {
		return nil, err
	}
	points := make([]types.OrderTrajectoryPoint, 0, len(records))
	for _, record := range records {
		reportTime := record.RecordedAt.Unix()
		if record.RecordedAt.IsZero() {
			reportTime = 0
		}
		points = append(points, types.OrderTrajectoryPoint{
			OrderID:    record.OrderID,
			DriverID:   record.DriverID,
			Longitude:  record.Longitude,
			Latitude:   record.Latitude,
			SpeedKmh:   record.SpeedKmh,
			Heading:    record.Heading,
			ReportTime: reportTime,
		})
	}
	return &types.GetOrderTrajectoryResponse{
		OrderID: orderID,
		Points:  points,
		Total:   int64(len(points)),
	}, nil
}

func canDriverViewOrder(driverID int64, order *orderproto.GetOrderResponse) bool {
	if driverID <= 0 || order == nil {
		return false
	}
	if order.GetDriverId() == driverID {
		return true
	}
	return order.GetDriverId() == 0 && order.GetStatus() == orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT
}

func (l *OrderLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}

func (l *OrderLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

func (l *OrderLogic) priceClient() (svc.PriceClient, error) {
	if l.svcCtx == nil || l.svcCtx.PriceClient == nil {
		return nil, ErrPriceClientNotConfigured
	}
	return l.svcCtx.PriceClient, nil
}

func (l *OrderLogic) dispatchClient() (svc.DispatchClient, error) {
	if l.svcCtx == nil || l.svcCtx.DispatchClient == nil {
		return nil, ErrDispatchClientNotConfigured
	}
	return l.svcCtx.DispatchClient, nil
}

// removeFromAvailable 从 driver:available:%d 集合移除指定订单，接单/拒单后调用，
// 避免大厅和 WS 在集合 TTL 过期前重复展示已处理的订单。
func (l *OrderLogic) removeFromAvailable(driverID, orderID int64) {
	if l.svcCtx == nil || l.svcCtx.RedisClient == nil || driverID <= 0 || orderID <= 0 {
		return
	}
	key := fmt.Sprintf(constants.RedisDriverAvailable, driverID)
	_ = l.svcCtx.RedisClient.SRem(l.ctx, key, fmt.Sprint(orderID)).Err()
}
