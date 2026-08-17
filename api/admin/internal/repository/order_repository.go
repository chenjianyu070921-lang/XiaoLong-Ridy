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

// ListAbnormal 查询后台异常订单列表。
// 异常来源包含订单取消、支付失败或退款、派单拒绝/超时/取消等信号，接口只做查询不修改订单状态。
func (r *OrderRepository) ListAbnormal(ctx context.Context, req types.AbnormalOrderListRequest) ([]types.AbnormalOrderDTO, int64, error) {
	where, args := buildAbnormalOrderWhere(req)
	countSQL := `
		SELECT COUNT(1)
		FROM ride_order o
	` + where
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := normalizeOffset(req.Page, req.PageSize)
	limit := normalizePageSize(req.PageSize)
	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, o.order_no, o.user_id, o.driver_id, o.car_type, o.from_address,
		       CAST(o.from_longitude AS CHAR), CAST(o.from_latitude AS CHAR), o.to_address,
		       CAST(o.to_longitude AS CHAR), CAST(o.to_latitude AS CHAR),
		       o.estimated_distance_m, o.estimated_duration_s, CAST(o.estimated_price AS CHAR),
		       o.status, o.cancel_reason, o.cancel_by, o.created_at, o.updated_at, o.deleted_at,
		       COALESCE((SELECT MAX(p.status) FROM payment p WHERE p.order_id = o.id), 0) AS payment_status,
		       COALESCE((SELECT MAX(d.status) FROM dispatch_record d WHERE d.order_id = o.id), 0) AS dispatch_status
		FROM ride_order o
	`+where+`
		ORDER BY o.id DESC
		LIMIT ? OFFSET ?
	`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]types.AbnormalOrderDTO, 0)
	for rows.Next() {
		item, err := scanAbnormalOrderRows(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *item)
	}
	return list, total, rows.Err()
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

// buildAbnormalOrderWhere 组装异常订单查询条件。
// 为避免重复 JOIN 导致行数膨胀，查询结果会在 ListAbnormal 中按订单 ID 聚合。
func buildAbnormalOrderWhere(req types.AbnormalOrderListRequest) (string, []any) {
	parts := []string{"o.deleted_at IS NULL"}
	args := make([]any, 0)

	switch req.AbnormalType {
	case "cancel":
		parts = append(parts, "(o.status = 6 OR o.cancel_reason <> '')")
	case "payment":
		parts = append(parts, "EXISTS (SELECT 1 FROM payment p WHERE p.order_id = o.id AND p.status IN (3, 4))")
	case "dispatch":
		parts = append(parts, "EXISTS (SELECT 1 FROM dispatch_record d WHERE d.order_id = o.id AND d.status IN (3, 4, 5))")
	default:
		parts = append(parts, `(
			(o.status = 6 OR o.cancel_reason <> '')
			OR EXISTS (SELECT 1 FROM payment p WHERE p.order_id = o.id AND p.status IN (3, 4))
			OR EXISTS (SELECT 1 FROM dispatch_record d WHERE d.order_id = o.id AND d.status IN (3, 4, 5))
		)`)
	}

	if req.Keyword != "" {
		parts = append(parts, "o.order_no LIKE ?")
		args = append(args, "%"+req.Keyword+"%")
	}
	if req.UserID > 0 {
		parts = append(parts, "o.user_id = ?")
		args = append(args, req.UserID)
	}
	if req.DriverID > 0 {
		parts = append(parts, "o.driver_id = ?")
		args = append(args, req.DriverID)
	}
	if req.StartTime != "" {
		parts = append(parts, "o.created_at >= ?")
		args = append(args, req.StartTime)
	}
	if req.EndTime != "" {
		parts = append(parts, "o.created_at <= ?")
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

// scanAbnormalOrderRows 扫描异常订单聚合查询结果。
// 该函数会补充异常类型和原因，前端可直接展示到异常订单列表。
func scanAbnormalOrderRows(rows *sql.Rows) (*types.AbnormalOrderDTO, error) {
	order, paymentStatus, dispatchStatus, err := scanOrderWithStatuses(rows)
	if err != nil {
		return nil, err
	}
	dto := types.AbnormalOrderDTO{
		OrderDTO:       orderToDTO(order),
		PaymentStatus:  paymentStatus,
		DispatchStatus: dispatchStatus,
	}
	dto.AbnormalType, dto.AbnormalReason = resolveAbnormalReason(order, paymentStatus, dispatchStatus)
	return &dto, nil
}

// scanOrderWithStatuses 扫描订单基础字段以及支付、派单状态。
// 该函数只服务异常订单聚合查询，避免影响原有订单列表扫描逻辑。
func scanOrderWithStatuses(rows *sql.Rows) (model.RideOrder, int32, int32, error) {
	var item model.RideOrder
	var deletedAt sql.NullTime
	var paymentStatus, dispatchStatus int32
	err := rows.Scan(
		&item.ID, &item.OrderNo, &item.UserID, &item.DriverID, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &item.CreatedAt, &item.UpdatedAt, &deletedAt,
		&paymentStatus, &dispatchStatus,
	)
	if err != nil {
		return item, 0, 0, fmt.Errorf("scan abnormal order row: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return item, paymentStatus, dispatchStatus, nil
}

// orderToDTO 将订单模型转换为异常订单接口内嵌的 DTO。
// 这里放在 repository 内部用于避免与 logic 包形成循环依赖。
func orderToDTO(order model.RideOrder) types.OrderDTO {
	return types.OrderDTO{
		ID:                 order.ID,
		OrderNo:            order.OrderNo,
		UserID:             order.UserID,
		DriverID:           order.DriverID,
		CarType:            order.CarType,
		FromAddress:        order.FromAddress,
		FromLongitude:      order.FromLongitude,
		FromLatitude:       order.FromLatitude,
		ToAddress:          order.ToAddress,
		ToLongitude:        order.ToLongitude,
		ToLatitude:         order.ToLatitude,
		EstimatedDistanceM: order.EstimatedDistanceM,
		EstimatedDurationS: order.EstimatedDurationS,
		EstimatedPrice:     order.EstimatedPrice,
		Status:             order.Status,
		CancelReason:       order.CancelReason,
		CancelBy:           order.CancelBy,
		CreatedAt:          FormatTime(order.CreatedAt),
		UpdatedAt:          FormatTime(order.UpdatedAt),
	}
}

// resolveAbnormalReason 根据订单、支付和派单状态推导异常类型与原因。
// 推导优先级为取消 > 支付 > 派单，保证一个订单在列表中只有一个主异常标签。
func resolveAbnormalReason(order model.RideOrder, paymentStatus, dispatchStatus int32) (string, string) {
	if order.Status == 6 || order.CancelReason != "" {
		if order.CancelReason != "" {
			return "cancel", order.CancelReason
		}
		return "cancel", "订单已取消"
	}
	if paymentStatus == 3 {
		return "payment", "支付失败"
	}
	if paymentStatus == 4 {
		return "payment", "订单存在退款记录"
	}
	if dispatchStatus == 3 {
		return "dispatch", "司机拒绝接单"
	}
	if dispatchStatus == 4 {
		return "dispatch", "派单超时"
	}
	if dispatchStatus == 5 {
		return "dispatch", "派单已取消"
	}
	return "unknown", "未知异常"
}
