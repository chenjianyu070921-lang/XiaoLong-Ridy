package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListOrdersLogic 处理订单列表查询 RPC。
type ListOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListOrdersLogic 创建订单列表逻辑对象。
func NewListOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrdersLogic {
	return &ListOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrders 查询订单主表列表。
func (l *ListOrdersLogic) ListOrders(in *adminsvc.OrderListRequest) (*adminsvc.OrderListResponse, error) {
	where, args := buildOrderWhere(in)
	var total int64
	if err := l.svcCtx.MySQL.QueryRowContext(l.ctx, `SELECT COUNT(1) FROM ride_order `+where, args...).Scan(&total); err != nil {
		return nil, err
	}
	limit := normalizePageSize(in.GetPageSize())
	queryArgs := append(args, limit, offset(in.GetPage(), in.GetPageSize()))
	rows, err := l.svcCtx.MySQL.QueryContext(l.ctx, `
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
		return nil, err
	}
	defer rows.Close()

	list := make([]*adminsvc.Order, 0)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return &adminsvc.OrderListResponse{List: list, Total: total, Page: normalizePage(in.GetPage()), PageSize: limit}, rows.Err()
}
