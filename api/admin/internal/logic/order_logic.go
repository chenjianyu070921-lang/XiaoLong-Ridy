package logic

import (
	"context"

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
	detail := &types.OrderDetailDTO{Order: orderPBToDTO(resp.Order)}
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
