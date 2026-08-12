package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// ErrOrderNotFound 表示没有找到指定订单。
var ErrOrderNotFound = errors.New("order not found")

// OrderRepository 封装订单监控相关数据访问。
// P0 阶段负责订单主表查询和详情聚合，不处理订单状态变更。
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository 创建订单仓储。
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// List 查询订单列表。
func (r *OrderRepository) List(ctx context.Context, req types.OrderListRequest) ([]model.RideOrder, int64, error) {
	where, args := buildOrderWhere(req)
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM ride_order `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_no, user_id, driver_id, car_type, from_address,
		       CAST(from_longitude AS CHAR), CAST(from_latitude AS CHAR), to_address,
		       CAST(to_longitude AS CHAR), CAST(to_latitude AS CHAR),
		       estimated_distance_m, estimated_duration_s, CAST(estimated_price AS CHAR),
		       status, cancel_reason, cancel_by, created_at, updated_at, deleted_at
		FROM ride_order `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]model.RideOrder, 0)
	for rows.Next() {
		item, err := scanOrderRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
}

// GetByID 查询订单主信息。
func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*model.RideOrder, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, order_no, user_id, driver_id, car_type, from_address,
		       CAST(from_longitude AS CHAR), CAST(from_latitude AS CHAR), to_address,
		       CAST(to_longitude AS CHAR), CAST(to_latitude AS CHAR),
		       estimated_distance_m, estimated_duration_s, CAST(estimated_price AS CHAR),
		       status, cancel_reason, cancel_by, created_at, updated_at, deleted_at
		FROM ride_order
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	var item model.RideOrder
	var deletedAt sql.NullTime
	err := row.Scan(
		&item.ID, &item.OrderNo, &item.UserID, &item.DriverID, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &item.CreatedAt, &item.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("scan order: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}

// ListStatusLogs 查询订单状态流转日志。
func (r *OrderRepository) ListStatusLogs(ctx context.Context, orderID int64) ([]types.OrderStatusLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, from_status, to_status, operator_type, operator_id, remark, created_at
		FROM order_status_log
		WHERE order_id = ?
		ORDER BY id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]types.OrderStatusLog, 0)
	for rows.Next() {
		var item types.OrderStatusLog
		var createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OrderID, &item.FromStatus, &item.ToStatus, &item.OperatorType, &item.OperatorID, &item.Remark, &createdAt); err != nil {
			return nil, err
		}
		item.CreatedAt = formatNullTime(createdAt)
		list = append(list, item)
	}
	return list, rows.Err()
}

// ListDispatchRecords 查询订单派单记录。
func (r *OrderRepository) ListDispatchRecords(ctx context.Context, orderID int64) ([]types.DispatchRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, driver_id, dispatch_type, status, CAST(match_score AS CHAR), remark, created_at, updated_at
		FROM dispatch_record
		WHERE order_id = ?
		ORDER BY id ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]types.DispatchRecord, 0)
	for rows.Next() {
		var item types.DispatchRecord
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.OrderID, &item.DriverID, &item.DispatchType, &item.Status, &item.MatchScore, &item.Remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		item.CreatedAt = formatNullTime(createdAt)
		item.UpdatedAt = formatNullTime(updatedAt)
		list = append(list, item)
	}
	return list, rows.Err()
}

// GetOrderPrice 查询订单价格明细，若不存在则返回 nil。
func (r *OrderRepository) GetOrderPrice(ctx context.Context, orderID int64) (*types.OrderPrice, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, order_id, price_rule_id, CAST(estimated_price AS CHAR), CAST(actual_price AS CHAR),
		       CAST(base_fee AS CHAR), CAST(distance_fee AS CHAR), CAST(time_fee AS CHAR),
		       CAST(night_fee AS CHAR), CAST(dynamic_fee AS CHAR), CAST(discount_amount AS CHAR),
		       CAST(platform_subsidy AS CHAR), CAST(payable_amount AS CHAR), status
		FROM order_price
		WHERE order_id = ?
	`, orderID)
	var item types.OrderPrice
	err := row.Scan(
		&item.ID, &item.OrderID, &item.PriceRuleID, &item.EstimatedPrice, &item.ActualPrice,
		&item.BaseFee, &item.DistanceFee, &item.TimeFee, &item.NightFee, &item.DynamicFee,
		&item.DiscountAmount, &item.PlatformSubsidy, &item.PayableAmount, &item.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &item, err
}

// GetPayment 查询订单支付单，若不存在则返回 nil。
func (r *OrderRepository) GetPayment(ctx context.Context, orderID int64) (*types.Payment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, payment_no, order_id, user_id, CAST(amount AS CHAR), channel, status,
		       transaction_id, CAST(refund_amount AS CHAR), paid_at
		FROM payment
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, orderID)
	var item types.Payment
	var paidAt sql.NullTime
	err := row.Scan(&item.ID, &item.PaymentNo, &item.OrderID, &item.UserID, &item.Amount, &item.Channel, &item.Status, &item.TransactionID, &item.RefundAmount, &paidAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	item.PaidAt = formatNullTime(paidAt)
	return &item, err
}

// GetSettlement 查询订单结算单，若不存在则返回 nil。
func (r *OrderRepository) GetSettlement(ctx context.Context, orderID int64) (*types.Settlement, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, settlement_no, order_id, driver_id, CAST(total_amount AS CHAR),
		       CAST(platform_commission_rate AS CHAR), CAST(platform_commission AS CHAR),
		       CAST(driver_income AS CHAR), status, settled_at
		FROM settlement
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, orderID)
	var item types.Settlement
	var settledAt sql.NullTime
	err := row.Scan(
		&item.ID, &item.SettlementNo, &item.OrderID, &item.DriverID, &item.TotalAmount,
		&item.PlatformCommissionRate, &item.PlatformCommission, &item.DriverIncome,
		&item.Status, &settledAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	item.SettledAt = formatNullTime(settledAt)
	return &item, err
}

// buildOrderWhere 组装订单列表筛选条件。
func buildOrderWhere(req types.OrderListRequest) (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if req.Keyword != "" {
		parts = append(parts, "order_no LIKE ?")
		args = append(args, "%"+req.Keyword+"%")
	}
	if req.Status > 0 {
		parts = append(parts, "status = ?")
		args = append(args, req.Status)
	}
	if req.UserID > 0 {
		parts = append(parts, "user_id = ?")
		args = append(args, req.UserID)
	}
	if req.DriverID > 0 {
		parts = append(parts, "driver_id = ?")
		args = append(args, req.DriverID)
	}
	if req.StartTime != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, req.EndTime)
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// scanOrderRows 从订单列表查询结果中扫描订单模型。
func scanOrderRows(rows *sql.Rows) (*model.RideOrder, error) {
	var item model.RideOrder
	var deletedAt sql.NullTime
	err := rows.Scan(
		&item.ID, &item.OrderNo, &item.UserID, &item.DriverID, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &item.CreatedAt, &item.UpdatedAt, &deletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan order row: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}
