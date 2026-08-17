package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrderStatusLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListOrderStatusLogsLogic 创建状态日志列表逻辑对象。
func NewListOrderStatusLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderStatusLogsLogic {
	return &ListOrderStatusLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListOrderStatusLogs 分页查询订单状态日志。
func (l *ListOrderStatusLogsLogic) ListOrderStatusLogs(in *proto.ListOrderStatusLogsRequest) (*proto.ListOrderStatusLogsResponse, error) {
	if in.OrderId <= 0 {
		return nil, ErrInvalidOrderParams
	}
	page := normalizePage(in.Page)
	pageSize := normalizePageSize(in.PageSize)

	list, total, err := l.svcCtx.OrderRepository.ListStatusLogs(l.ctx, uint64(in.OrderId), page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]*proto.OrderStatusLog, 0, len(list))
	for i := range list {
		item := list[i]
		items = append(items, &proto.OrderStatusLog{
			Id:           int64(item.Id),
			OrderId:      int64(item.OrderId),
			FromStatus:   int32(item.FromStatus),
			ToStatus:     int32(item.ToStatus),
			OperatorType: item.OperatorType,
			OperatorId:   int64(item.OperatorId),
			Remark:       item.Remark,
			CreatedAt:    item.CreatedAt.Unix(),
		})
	}

	return &proto.ListOrderStatusLogsResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
