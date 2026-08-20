package adminservicelogic

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
)

// buildOrderWhere 组装订单列表筛选条件。
func buildOrderWhere(in *adminsvc.OrderListRequest) (string, []any) {
	parts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if in.GetKeyword() != "" {
		parts = append(parts, "order_no LIKE ?")
		args = append(args, "%"+in.GetKeyword()+"%")
	}
	if in.GetStatus() > 0 {
		parts = append(parts, "status = ?")
		args = append(args, in.GetStatus())
	}
	if in.GetUserId() > 0 {
		parts = append(parts, "user_id = ?")
		args = append(args, in.GetUserId())
	}
	if in.GetDriverId() > 0 {
		parts = append(parts, "driver_id = ?")
		args = append(args, in.GetDriverId())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// scanOrder 将订单主表行转换为 protobuf 订单对象。
func scanOrder(rows *sql.Rows) (*adminsvc.Order, error) {
	var item adminsvc.Order
	var createdAt, updatedAt, deletedAt sql.NullTime
	if err := rows.Scan(
		&item.Id, &item.OrderNo, &item.UserId, &item.DriverId, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &createdAt, &updatedAt, &deletedAt,
	); err != nil {
		return nil, fmt.Errorf("scan orderclient row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	_ = deletedAt
	return &item, nil
}

// scanOrderStatusLog 将订单状态流转记录转成 protobuf。
func scanOrderStatusLog(rows *sql.Rows) (*adminsvc.OrderStatusLog, error) {
	var item adminsvc.OrderStatusLog
	var createdAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.OrderId, &item.FromStatus, &item.ToStatus, &item.OperatorType, &item.OperatorId, &item.Remark, &createdAt); err != nil {
		return nil, fmt.Errorf("scan orderclient status log row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	return &item, nil
}

// scanDispatchRecord 将派单记录转换为 protobuf。
func scanDispatchRecord(rows *sql.Rows) (*adminsvc.DispatchRecord, error) {
	var item adminsvc.DispatchRecord
	var createdAt, updatedAt sql.NullTime
	if err := rows.Scan(&item.Id, &item.OrderId, &item.DriverId, &item.DispatchType, &item.Status, &item.MatchScore, &item.Remark, &createdAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("scan dispatch record row: %w", err)
	}
	item.CreatedAt = formatNullTime(createdAt)
	item.UpdatedAt = formatNullTime(updatedAt)
	return &item, nil
}

// scanOrderPrice 将价格明细转换为 protobuf。
func scanOrderPrice(row *sql.Row) (*adminsvc.OrderPrice, error) {
	var item adminsvc.OrderPrice
	err := row.Scan(
		&item.Id, &item.OrderId, &item.PriceRuleId, &item.EstimatedPrice, &item.ActualPrice,
		&item.BaseFee, &item.DistanceFee, &item.TimeFee, &item.NightFee, &item.DynamicFee,
		&item.DiscountAmount, &item.PlatformSubsidy, &item.PayableAmount, &item.Status,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan orderclient price: %w", err)
	}
	return &item, nil
}

// scanPayment 将支付记录转换为 protobuf。
func scanPayment(row *sql.Row) (*adminsvc.Payment, error) {
	var item adminsvc.Payment
	var paidAt sql.NullTime
	err := row.Scan(&item.Id, &item.PaymentNo, &item.OrderId, &item.UserId, &item.Amount, &item.Channel, &item.Status, &item.TransactionId, &item.RefundAmount, &paidAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	item.PaidAt = formatNullTime(paidAt)
	return &item, nil
}

// scanSettlement 将结算记录转换为 protobuf。
func scanSettlement(row *sql.Row) (*adminsvc.Settlement, error) {
	var item adminsvc.Settlement
	var settledAt sql.NullTime
	err := row.Scan(&item.Id, &item.SettlementNo, &item.OrderId, &item.DriverId, &item.TotalAmount, &item.PlatformCommissionRate, &item.PlatformCommission, &item.DriverIncome, &item.Status, &settledAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan settlement: %w", err)
	}
	item.SettledAt = formatNullTime(settledAt)
	return &item, nil
}
