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
	if l.svcCtx == nil || l.svcCtx.MySQL == nil {
		degraded = append(degraded, "price", "payment", "settlement")
	} else {
		price, err = scanOrderPrice(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
			SELECT id, order_id, price_rule_id, CAST(estimated_price AS CHAR), CAST(actual_price AS CHAR),
			       CAST(base_fee AS CHAR), CAST(distance_fee AS CHAR), CAST(time_fee AS CHAR),
			       CAST(night_fee AS CHAR), CAST(dynamic_fee AS CHAR), CAST(discount_amount AS CHAR),
			       CAST(platform_subsidy AS CHAR), CAST(payable_amount AS CHAR), status
			FROM order_price
			WHERE order_id = ?
		`, in.GetId()))
		if err != nil {
			degraded = append(degraded, "price")
			price = nil
		}

		payment, err = scanPayment(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
			SELECT id, payment_no, order_id, user_id, CAST(amount AS CHAR), channel, status,
			       transaction_id, CAST(refund_amount AS CHAR), paid_at
			FROM payment
			WHERE order_id = ?
			ORDER BY id DESC
			LIMIT 1
		`, in.GetId()))
		if err != nil {
			degraded = append(degraded, "payment")
			payment = nil
		}

		settlement, err = scanSettlement(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
			SELECT id, settlement_no, order_id, driver_id, CAST(total_amount AS CHAR),
			       CAST(platform_commission_rate AS CHAR), CAST(platform_commission AS CHAR),
			       CAST(driver_income AS CHAR), status, settled_at
			FROM settlement
			WHERE order_id = ?
			ORDER BY id DESC
			LIMIT 1
		`, in.GetId()))
		if err != nil {
			degraded = append(degraded, "settlement")
			settlement = nil
		}
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
