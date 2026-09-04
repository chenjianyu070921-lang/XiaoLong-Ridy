package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// OrderLogic 负责管理后台订单查询的 HTTP 到 RPC 适配。
type OrderLogic struct {
	ctx *svc.ServiceContext
}

// NewOrderLogic 创建订单逻辑。
func NewOrderLogic(ctx *svc.ServiceContext) *OrderLogic {
	return &OrderLogic{ctx: ctx}
}

// List 查询全量订单列表。
func (l *OrderLogic) List(ctx context.Context, req types.OrderListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListOrders(ctx, &adminclient.OrderListRequest{
		Page:      int32(req.Page),
		PageSize:  int32(req.PageSize),
		Keyword:   req.Keyword,
		Status:    req.Status,
		UserId:    req.UserID,
		DriverId:  req.DriverID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.OrderDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, orderPBToDTO(item))
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// ListAbnormal 查询异常订单列表。
func (l *OrderLogic) ListAbnormal(ctx context.Context, req types.AbnormalOrderListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListAbnormalOrders(ctx, &adminclient.AbnormalOrderListRequest{
		Page:         int32(req.Page),
		PageSize:     int32(req.PageSize),
		Keyword:      req.Keyword,
		AbnormalType: req.AbnormalType,
		UserId:       req.UserID,
		DriverId:     req.DriverID,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.AbnormalOrderDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.AbnormalOrderDTO{
			OrderDTO: types.OrderDTO{
				ID:                 item.Id,
				OrderNo:            item.OrderNo,
				UserID:             item.UserId,
				DriverID:           item.DriverId,
				CarType:            item.CarType,
				FromAddress:        item.FromAddress,
				FromLongitude:      item.FromLongitude,
				FromLatitude:       item.FromLatitude,
				ToAddress:          item.ToAddress,
				ToLongitude:        item.ToLongitude,
				ToLatitude:         item.ToLatitude,
				EstimatedDistanceM: item.EstimatedDistanceM,
				EstimatedDurationS: item.EstimatedDurationS,
				EstimatedPrice:     item.EstimatedPrice,
				Status:             item.Status,
				CancelReason:       item.CancelReason,
				CancelBy:           item.CancelBy,
				CreatedAt:          item.CreatedAt,
				UpdatedAt:          item.UpdatedAt,
			},
			AbnormalType:   item.AbnormalType,
			AbnormalReason: item.AbnormalReason,
			PaymentStatus:  item.PaymentStatus,
			DispatchStatus: item.DispatchStatus,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// Detail 查询订单详情聚合信息。
func (l *OrderLogic) Detail(ctx context.Context, id int64) (*types.OrderDetailDTO, error) {
	resp, err := l.ctx.AdminSvc.GetOrder(ctx, &adminclient.OrderDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	detail := &types.OrderDetailDTO{Order: orderPBToDTO(resp.Order), Degraded: resp.GetDegraded()}
	for _, item := range resp.StatusLogs {
		detail.StatusLogs = append(detail.StatusLogs, types.OrderStatusLog{
			ID: item.Id, OrderID: item.OrderId, FromStatus: item.FromStatus, ToStatus: item.ToStatus,
			OperatorType: item.OperatorType, OperatorID: item.OperatorId, Remark: item.Remark, CreatedAt: item.CreatedAt,
		})
	}
	for _, item := range resp.DispatchRecords {
		detail.DispatchRecords = append(detail.DispatchRecords, types.DispatchRecord{
			ID: item.Id, OrderID: item.OrderId, DriverID: item.DriverId, DispatchType: item.DispatchType,
			Status: item.Status, MatchScore: item.MatchScore, Remark: item.Remark, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	if resp.Price != nil {
		detail.Price = &types.OrderPrice{
			ID: resp.Price.Id, OrderID: resp.Price.OrderId, PriceRuleID: resp.Price.PriceRuleId,
			EstimatedPrice: resp.Price.EstimatedPrice, ActualPrice: resp.Price.ActualPrice,
			BaseFee: resp.Price.BaseFee, DistanceFee: resp.Price.DistanceFee, TimeFee: resp.Price.TimeFee,
			NightFee: resp.Price.NightFee, DynamicFee: resp.Price.DynamicFee, DiscountAmount: resp.Price.DiscountAmount,
			PlatformSubsidy: resp.Price.PlatformSubsidy, PayableAmount: resp.Price.PayableAmount, Status: resp.Price.Status,
		}
	}
	if resp.Payment != nil {
		detail.Payment = &types.Payment{
			ID: resp.Payment.Id, PaymentNo: resp.Payment.PaymentNo, OrderID: resp.Payment.OrderId, UserID: resp.Payment.UserId,
			Amount: resp.Payment.Amount, Channel: resp.Payment.Channel, Status: resp.Payment.Status,
			TransactionID: resp.Payment.TransactionId, RefundAmount: resp.Payment.RefundAmount, PaidAt: resp.Payment.PaidAt,
		}
	}
	if resp.Settlement != nil {
		detail.Settlement = &types.Settlement{
			ID: resp.Settlement.Id, SettlementNo: resp.Settlement.SettlementNo, OrderID: resp.Settlement.OrderId,
			DriverID: resp.Settlement.DriverId, TotalAmount: resp.Settlement.TotalAmount,
			PlatformCommissionRate: resp.Settlement.PlatformCommissionRate, PlatformCommission: resp.Settlement.PlatformCommission,
			DriverIncome: resp.Settlement.DriverIncome, Status: resp.Settlement.Status, SettledAt: resp.Settlement.SettledAt,
		}
	}
	return detail, nil
}

// Track 查询订单轨迹点。
// HTTP 层只接收查询参数，轨迹数据由 adminsvc 继续转发到 locationsvc，避免后台直接读取位置服务表。
func (l *OrderLogic) Track(ctx context.Context, req types.OrderTrackRequest) (*types.OrderTrackDTO, error) {
	resp, err := l.ctx.AdminSvc.GetOrderTrack(ctx, &adminclient.OrderTrackRequest{
		OrderId:   req.OrderID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
	})
	if err != nil {
		return nil, err
	}
	points := make([]types.OrderTrackPointDTO, 0, len(resp.GetPoints()))
	for _, item := range resp.GetPoints() {
		points = append(points, types.OrderTrackPointDTO{
			ID:         item.GetId(),
			OrderID:    item.GetOrderId(),
			DriverID:   item.GetDriverId(),
			Longitude:  item.GetLongitude(),
			Latitude:   item.GetLatitude(),
			SpeedKmh:   item.GetSpeedKmh(),
			Direction:  item.GetDirection(),
			RecordedAt: item.GetRecordedAt(),
		})
	}
	return &types.OrderTrackDTO{Points: points}, nil
}

// Cancel 调用 adminsvc 取消订单，HTTP 层只负责传递管理员、订单和取消原因。
func (l *OrderLogic) Cancel(ctx context.Context, id int64, req types.OrderCancelRequest, session *model.AdminSession, ip string) error {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "后台取消订单"
	}
	_, err := l.ctx.AdminSvc.CancelOrder(ctx, &adminclient.AdminCancelOrderRequest{
		OrderId:   id,
		Reason:    reason,
		AdminId:   session.AdminID,
		Ip:        ip,
		RequestId: strings.TrimSpace(req.RequestID),
	})
	return err
}

// Redispatch 调用 adminsvc 执行后台人工改派。
// HTTP 网关只透传管理员会话、目标司机和幂等号，订单状态校验及派单触发由下游 ordersvc 完成。
func (l *OrderLogic) Redispatch(ctx context.Context, id int64, req types.OrderRedispatchRequest, session *model.AdminSession, ip string) (*types.OrderRedispatchResponse, error) {
	resp, err := l.ctx.AdminSvc.RedispatchOrder(ctx, &adminclient.AdminRedispatchOrderRequest{
		OrderId:     id,
		NewDriverId: req.NewDriverID,
		Reason:      strings.TrimSpace(req.Reason),
		AdminId:     session.AdminID,
		Ip:          ip,
		RequestId:   strings.TrimSpace(req.RequestID),
	})
	if err != nil {
		return nil, err
	}
	return &types.OrderRedispatchResponse{
		OrderID:  resp.GetOrderId(),
		Status:   resp.GetStatus(),
		DriverID: resp.GetDriverId(),
		Message:  resp.GetMessage(),
	}, nil
}

// Refund 调用 adminsvc 执行后台订单退款。
// request_id 会在 adminsvc 层作为 refund_no 传入 ordersvc，确保资金类操作具备可追踪幂等号。
func (l *OrderLogic) Refund(ctx context.Context, id int64, req types.OrderRefundRequest, session *model.AdminSession, ip string) (*types.OrderRefundResponse, error) {
	resp, err := l.ctx.AdminSvc.RefundOrder(ctx, &adminclient.AdminRefundOrderRequest{
		OrderId:           id,
		RefundAmountCents: req.RefundAmountCents,
		Reason:            strings.TrimSpace(req.Reason),
		AdminId:           session.AdminID,
		Ip:                ip,
		RequestId:         strings.TrimSpace(req.RequestID),
	})
	if err != nil {
		return nil, err
	}
	return &types.OrderRefundResponse{
		OrderID:     resp.GetOrderId(),
		Status:      resp.GetStatus(),
		RefundCents: resp.GetRefundCents(),
		RefundNo:    resp.GetRefundNo(),
		Message:     resp.GetMessage(),
	}, nil
}

// orderPBToDTO 将订单 protobuf 对象转换为 HTTP DTO。
func orderPBToDTO(item *adminclient.Order) types.OrderDTO {
	if item == nil {
		return types.OrderDTO{}
	}
	return types.OrderDTO{
		ID: item.Id, OrderNo: item.OrderNo, UserID: item.UserId, DriverID: item.DriverId, CarType: item.CarType,
		FromAddress: item.FromAddress, FromLongitude: item.FromLongitude, FromLatitude: item.FromLatitude,
		ToAddress: item.ToAddress, ToLongitude: item.ToLongitude, ToLatitude: item.ToLatitude,
		EstimatedDistanceM: item.EstimatedDistanceM, EstimatedDurationS: item.EstimatedDurationS,
		EstimatedPrice: item.EstimatedPrice, Status: item.Status, CancelReason: item.CancelReason,
		CancelBy: item.CancelBy, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
}
