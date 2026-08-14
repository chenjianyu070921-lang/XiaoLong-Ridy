package adminservicelogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
)

// ListAbnormalOrdersLogic 处理后台异常订单查询。
type ListAbnormalOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewListAbnormalOrdersLogic 创建异常订单查询逻辑。
func NewListAbnormalOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAbnormalOrdersLogic {
	return &ListAbnormalOrdersLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListAbnormalOrders 查询取消、支付失败和派单异常订单。
func (l *ListAbnormalOrdersLogic) ListAbnormalOrders(in *adminsvc.AbnormalOrderListRequest) (*adminsvc.AbnormalOrderListResponse, error) {
	where, args := buildAbnormalOrderWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM ride_order o `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
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
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.AbnormalOrder, 0)
	for rows.Next() {
		item, err := scanAbnormalOrderRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.AbnormalOrderListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}

// buildAbnormalOrderWhere 组装异常订单筛选条件。
func buildAbnormalOrderWhere(in *adminsvc.AbnormalOrderListRequest) (string, []any) {
	parts := []string{"o.deleted_at IS NULL"}
	args := make([]any, 0)

	switch in.GetAbnormalType() {
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

	if in.GetKeyword() != "" {
		parts = append(parts, "o.order_no LIKE ?")
		args = append(args, "%"+in.GetKeyword()+"%")
	}
	if in.GetUserId() > 0 {
		parts = append(parts, "o.user_id = ?")
		args = append(args, in.GetUserId())
	}
	if in.GetDriverId() > 0 {
		parts = append(parts, "o.driver_id = ?")
		args = append(args, in.GetDriverId())
	}
	if in.GetStartTime() != "" {
		parts = append(parts, "o.created_at >= ?")
		args = append(args, in.GetStartTime())
	}
	if in.GetEndTime() != "" {
		parts = append(parts, "o.created_at <= ?")
		args = append(args, in.GetEndTime())
	}
	return "WHERE " + strings.Join(parts, " AND "), args
}

// scanAbnormalOrderRow 扫描异常订单聚合结果。
func scanAbnormalOrderRow(rows *sql.Rows) (*adminsvc.AbnormalOrder, error) {
	var item adminsvc.AbnormalOrder
	var paymentStatus, dispatchStatus int32
	if err := rows.Scan(
		&item.Id, &item.OrderNo, &item.UserId, &item.DriverId, &item.CarType, &item.FromAddress,
		&item.FromLongitude, &item.FromLatitude, &item.ToAddress, &item.ToLongitude, &item.ToLatitude,
		&item.EstimatedDistanceM, &item.EstimatedDurationS, &item.EstimatedPrice,
		&item.Status, &item.CancelReason, &item.CancelBy, &item.CreatedAt, &item.UpdatedAt, new(sql.NullTime),
		&paymentStatus, &dispatchStatus,
	); err != nil {
		return nil, fmt.Errorf("scan abnormal order row: %w", err)
	}
	item.PaymentStatus = paymentStatus
	item.DispatchStatus = dispatchStatus
	item.AbnormalType, item.AbnormalReason = resolveAbnormalReason(item.Status, item.CancelReason, paymentStatus, dispatchStatus)
	return &item, nil
}

// resolveAbnormalReason 根据异常来源推断异常类型和原因。
func resolveAbnormalReason(orderStatus int32, cancelReason string, paymentStatus, dispatchStatus int32) (string, string) {
	if orderStatus == 6 || cancelReason != "" {
		if cancelReason != "" {
			return "cancel", cancelReason
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
