package adminservicelogic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	ordersvc "XiaoLong-Ridy/rpc/ordersvc/proto"

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
	// 可由 ordersvc 完整表达的筛选条件优先查询真实订单服务。
	// 订单号关键字和创建时间范围暂未纳入 RPC 契约，存在这些条件时保留兼容查询。
	if l.svcCtx != nil && l.svcCtx.OrdersSvc != nil && orderListCanUseOrdersRPC(in) {
		return l.listOrdersFromRPC(in)
	}

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

// orderListCanUseOrdersRPC 判断后台筛选条件是否完全落在 ordersvc.ListOrders 的能力范围内。
func orderListCanUseOrdersRPC(in *adminsvc.OrderListRequest) bool {
	if in == nil {
		return false
	}
	return strings.TrimSpace(in.GetKeyword()) == "" &&
		strings.TrimSpace(in.GetStartTime()) == "" &&
		strings.TrimSpace(in.GetEndTime()) == ""
}

// listOrdersFromRPC 通过 ordersvc 查询真实订单列表，并转换为管理后台响应结构。
func (l *ListOrdersLogic) listOrdersFromRPC(in *adminsvc.OrderListRequest) (*adminsvc.OrderListResponse, error) {
	resp, err := l.svcCtx.OrdersSvc.ListOrders(l.ctx, &ordersvc.ListOrdersRequest{
		UserId: in.GetUserId(), DriverId: in.GetDriverId(), Status: ordersvc.OrderStatus(in.GetStatus()),
		Page: in.GetPage(), PageSize: in.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	list := make([]*adminsvc.Order, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, orderSummaryFromOrdersRPC(item))
	}
	return &adminsvc.OrderListResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// orderSummaryFromOrdersRPC 将 ordersvc 订单摘要转换为管理后台订单列表项。
// 列表字段以 ordersvc 返回为准，避免全量查询时用筛选条件反推 user_id/driver_id。
func orderSummaryFromOrdersRPC(item *ordersvc.OrderSummary) *adminsvc.Order {
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
