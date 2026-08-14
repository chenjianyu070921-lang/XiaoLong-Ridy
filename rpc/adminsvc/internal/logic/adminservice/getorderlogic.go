package adminservicelogic

import (
	"context"
	"database/sql"
	"errors"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

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
	row := l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, order_no, user_id, driver_id, car_type, from_address,
		       CAST(from_longitude AS CHAR), CAST(from_latitude AS CHAR), to_address,
		       CAST(to_longitude AS CHAR), CAST(to_latitude AS CHAR),
		       estimated_distance_m, estimated_duration_s, CAST(estimated_price AS CHAR),
		       status, cancel_reason, cancel_by, created_at, updated_at, deleted_at
		FROM ride_order
		WHERE id = ? AND deleted_at IS NULL
	`, in.GetId())
	order, err := scanOrderSingle(row)
	if err != nil {
		return nil, err
	}

	statusRows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, order_id, from_status, to_status, operator_type, operator_id, remark, created_at
		FROM order_status_log
		WHERE order_id = ?
		ORDER BY id ASC
	`, in.GetId())
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()
	statusLogs := make([]*adminsvc.OrderStatusLog, 0)
	for statusRows.Next() {
		item, err := scanOrderStatusLog(statusRows)
		if err != nil {
			return nil, err
		}
		statusLogs = append(statusLogs, item)
	}

	dispatchRows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
		SELECT id, order_id, driver_id, dispatch_type, status, CAST(match_score AS CHAR), remark, created_at, updated_at
		FROM dispatch_record
		WHERE order_id = ?
		ORDER BY id ASC
	`, in.GetId())
	if err != nil {
		return nil, err
	}
	defer dispatchRows.Close()
	dispatchRecords := make([]*adminsvc.DispatchRecord, 0)
	for dispatchRows.Next() {
		item, err := scanDispatchRecord(dispatchRows)
		if err != nil {
			return nil, err
		}
		dispatchRecords = append(dispatchRecords, item)
	}

	price, err := scanOrderPrice(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, order_id, price_rule_id, CAST(estimated_price AS CHAR), CAST(actual_price AS CHAR),
		       CAST(base_fee AS CHAR), CAST(distance_fee AS CHAR), CAST(time_fee AS CHAR),
		       CAST(night_fee AS CHAR), CAST(dynamic_fee AS CHAR), CAST(discount_amount AS CHAR),
		       CAST(platform_subsidy AS CHAR), CAST(payable_amount AS CHAR), status
		FROM order_price
		WHERE order_id = ?
	`, in.GetId()))
	if err != nil {
		return nil, err
	}

	payment, err := scanPayment(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, payment_no, order_id, user_id, CAST(amount AS CHAR), channel, status,
		       transaction_id, CAST(refund_amount AS CHAR), paid_at
		FROM payment
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, in.GetId()))
	if err != nil {
		return nil, err
	}

	settlement, err := scanSettlement(l.svcCtx.MySQL.QueryRowContext(l.ctx, `
		SELECT id, settlement_no, order_id, driver_id, CAST(total_amount AS CHAR),
		       CAST(platform_commission_rate AS CHAR), CAST(platform_commission AS CHAR),
		       CAST(driver_income AS CHAR), status, settled_at
		FROM settlement
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, in.GetId()))
	if err != nil {
		return nil, err
	}

	return &adminsvc.OrderDetail{
		Order:           order,
		StatusLogs:      statusLogs,
		DispatchRecords: dispatchRecords,
		Price:           price,
		Payment:         payment,
		Settlement:      settlement,
	}, nil
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
