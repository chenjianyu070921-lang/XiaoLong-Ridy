package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	dispatchsvc "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	ordersvc "XiaoLong-Ridy/rpc/ordersvc/proto"
	paysvcproto "XiaoLong-Ridy/rpc/paysvc/proto"
	pricesvcproto "XiaoLong-Ridy/rpc/pricesvc/price"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetOrderLogic 处理订单详情查询 RPC。
type GetOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetOrderLogic 创建订单详情逻辑对象。
func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrder 查询订单详情并聚合状态、派单、价格、支付和结算信息。
func (l *GetOrderLogic) GetOrder(in *adminsvc.OrderDetailRequest) (*adminsvc.OrderDetail, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "订单ID不能为空")
	}
	order, err := l.getOrderMain(in.GetId())
	if err != nil {
		return nil, err
	}

	degraded := make([]string, 0)
	statusLogs := make([]*adminsvc.OrderStatusLog, 0)
	if l.svcCtx.OrdersSvc != nil {
		statusResp, statusErr := l.svcCtx.OrdersSvc.ListOrderStatusLogs(l.ctx, &ordersvc.ListOrderStatusLogsRequest{
			OrderId:  in.GetId(),
			Page:     1,
			PageSize: 100,
		})
		if statusErr != nil {
			degraded = append(degraded, "status_logs")
		} else {
			for _, item := range statusResp.GetList() {
				statusLogs = append(statusLogs, &adminsvc.OrderStatusLog{
					Id:           item.GetId(),
					OrderId:      item.GetOrderId(),
					FromStatus:   item.GetFromStatus(),
					ToStatus:     item.GetToStatus(),
					OperatorType: item.GetOperatorType(),
					OperatorId:   item.GetOperatorId(),
					Remark:       item.GetRemark(),
					CreatedAt:    unixText(item.GetCreatedAt()),
				})
			}
		}
	}

	dispatchRecords := make([]*adminsvc.DispatchRecord, 0)
	if l.svcCtx.DispatchSvc != nil {
		dispatchResp, dispatchErr := l.svcCtx.DispatchSvc.ListDispatchRecords(l.ctx, &dispatchsvc.ListDispatchRecordsRequest{
			OrderId:  in.GetId(),
			Page:     1,
			PageSize: 100,
		})
		if dispatchErr != nil {
			degraded = append(degraded, "dispatch_records")
		} else {
			for _, item := range dispatchResp.GetList() {
				dispatchRecords = append(dispatchRecords, &adminsvc.DispatchRecord{
					Id:           item.GetId(),
					OrderId:      item.GetOrderId(),
					DriverId:     item.GetDriverId(),
					DispatchType: item.GetDispatchType(),
					Status:       item.GetStatus(),
					MatchScore:   fmt.Sprintf("%.6f", item.GetMatchScore()),
					Remark:       item.GetRemark(),
					CreatedAt:    unixText(item.GetCreatedAt()),
					UpdatedAt:    unixText(item.GetUpdatedAt()),
				})
			}
		}
	}

	price, payment, settlement := (*adminsvc.OrderPrice)(nil), (*adminsvc.Payment)(nil), (*adminsvc.Settlement)(nil)
	// 价格明细改由 pricesvc 查询，支付与结算改由 paysvc 查询，避免后台跨服务直读资金与价格表。
	if l.svcCtx != nil && l.svcCtx.PricesSvc != nil {
		if priceResp, priceErr := l.svcCtx.PricesSvc.GetOrderPrice(l.ctx, &pricesvcproto.GetOrderPriceRequest{OrderId: in.GetId()}); priceErr != nil {
			degraded = append(degraded, "price")
		} else if priceResp != nil {
			price = &adminsvc.OrderPrice{
				Id:              priceResp.GetId(),
				OrderId:         priceResp.GetOrderId(),
				PriceRuleId:     priceResp.GetPriceRuleId(),
				EstimatedPrice:  formatCents(priceResp.GetEstimatedPriceCents()),
				ActualPrice:     formatCents(priceResp.GetActualPriceCents()),
				BaseFee:         formatCents(priceResp.GetBaseFeeCents()),
				DistanceFee:     formatCents(priceResp.GetDistanceFeeCents()),
				TimeFee:         formatCents(priceResp.GetTimeFeeCents()),
				NightFee:        formatCents(priceResp.GetNightFeeCents()),
				DynamicFee:      formatCents(priceResp.GetDynamicFeeCents()),
				DiscountAmount:  formatCents(priceResp.GetDiscountAmountCents()),
				PlatformSubsidy: formatCents(priceResp.GetPlatformSubsidyCents()),
				PayableAmount:   formatCents(priceResp.GetPayableAmountCents()),
				Status:          priceResp.GetStatus(),
			}
		}
	} else {
		degraded = append(degraded, "price")
	}

	if l.svcCtx != nil && l.svcCtx.PaySvc != nil {
		if payResp, payErr := l.svcCtx.PaySvc.GetPayment(l.ctx, &paysvcproto.GetPaymentRequest{OrderId: in.GetId()}); payErr != nil {
			degraded = append(degraded, "payment")
		} else if payResp != nil {
			payment = &adminsvc.Payment{
				Id:            payResp.GetPaymentId(),
				PaymentNo:     payResp.GetPaymentNo(),
				OrderId:       payResp.GetOrderId(),
				UserId:        payResp.GetUserId(),
				Amount:        formatCents(payResp.GetAmountCents()),
				Channel:       payResp.GetChannel(),
				Status:        payResp.GetStatus(),
				TransactionId: payResp.GetTransactionId(),
				RefundAmount:  formatCents(payResp.GetRefundAmountCents()),
				PaidAt:        unixText(payResp.GetPaidAt()),
			}
		}
		if setResp, setErr := l.svcCtx.PaySvc.GetSettlement(l.ctx, &paysvcproto.GetSettlementRequest{OrderId: in.GetId()}); setErr != nil {
			degraded = append(degraded, "settlement")
		} else if setResp != nil {
			settlement = &adminsvc.Settlement{
				Id:                     setResp.GetSettlementId(),
				SettlementNo:           setResp.GetSettlementNo(),
				OrderId:                setResp.GetOrderId(),
				DriverId:               setResp.GetDriverId(),
				TotalAmount:            formatCents(setResp.GetTotalAmountCents()),
				PlatformCommissionRate:  setResp.GetPlatformCommissionRate(),
				PlatformCommission:      formatCents(setResp.GetPlatformCommissionCents()),
				DriverIncome:            formatCents(setResp.GetDriverIncomeCents()),
				Status:                  setResp.GetStatus(),
				SettledAt:               unixText(setResp.GetSettledAt()),
			}
		}
	} else {
		degraded = append(degraded, "payment", "settlement")
	}

	return &adminsvc.OrderDetail{
		Order:           order,
		StatusLogs:      statusLogs,
		DispatchRecords: dispatchRecords,
		Price:           price,
		Payment:         payment,
		Settlement:      settlement,
		Degraded:        degraded,
	}, nil
}

// getOrderMain 查询订单主信息。
// 有 ordersvc client 时以订单服务返回为准；无下游 client 的最小本地模式才走兼容 SQL 只读查询。
func (l *GetOrderLogic) getOrderMain(orderID int64) (*adminsvc.Order, error) {
	if l.svcCtx != nil && l.svcCtx.OrdersSvc != nil {
		resp, err := l.svcCtx.OrdersSvc.GetOrder(l.ctx, &ordersvc.GetOrderRequest{OrderId: orderID})
		if err != nil {
			return nil, err
		}
		return orderDetailFromOrdersRPC(resp), nil
	}
	row := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, order_no, user_id, driver_id, car_type, from_address,
		       CAST(from_longitude AS CHAR), CAST(from_latitude AS CHAR), to_address,
		       CAST(to_longitude AS CHAR), CAST(to_latitude AS CHAR),
		       estimated_distance_m, estimated_duration_s, CAST(estimated_price AS CHAR),
		       status, cancel_reason, cancel_by, created_at, updated_at, deleted_at
		FROM ride_order
		WHERE id = ? AND deleted_at IS NULL
	`, orderID)
	return scanOrderSingle(row)
}

// orderDetailFromOrdersRPC 将 ordersvc 订单详情转换为管理后台订单主信息。
func orderDetailFromOrdersRPC(item *ordersvc.GetOrderResponse) *adminsvc.Order {
	if item == nil {
		return nil
	}
	return &adminsvc.Order{
		Id:                 item.GetOrderId(),
		OrderNo:            item.GetOrderNo(),
		UserId:             item.GetUserId(),
		DriverId:           item.GetDriverId(),
		CarType:            item.GetCarType(),
		FromAddress:        item.GetFromAddress(),
		FromLongitude:      floatText(item.GetFromLongitude()),
		FromLatitude:       floatText(item.GetFromLatitude()),
		ToAddress:          item.GetToAddress(),
		ToLongitude:        floatText(item.GetToLongitude()),
		ToLatitude:         floatText(item.GetToLatitude()),
		EstimatedDistanceM: item.GetEstimatedDistanceM(),
		EstimatedDurationS: item.GetEstimatedDurationS(),
		EstimatedPrice:     formatCents(item.GetEstimatedPriceCents()),
		Status:             int32(item.GetStatus()),
		CancelReason:       item.GetCancelReason(),
		CancelBy:           item.GetCancelBy(),
		CreatedAt:          unixText(item.GetCreatedAt()),
		UpdatedAt:          unixText(item.GetUpdatedAt()),
	}
}

// scanOrderSingle 处理订单详情单行结果。
func scanOrderSingle(row *sql.Row) (*adminsvc.Order, error) {
	var item adminsvc.Order
	var createdAt, updatedAt, deletedAt sql.NullTime
	err := row.Scan(
		&item.Id, &item.OrderNo, &item.UserId, &item.DriverId, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &createdAt, &updatedAt, &deletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "订单不存在")
	}
	if err != nil {
		return nil, err
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	_ = deletedAt
	return &item, nil
}
