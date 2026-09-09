package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/dispatchsvc/internal/svc"
	"XiaoLong-Ridy/rpc/dispatchsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultDispatchTimeoutSeconds = 60
	maxTimeoutOrderPageSize       = 100
)

type ListTimeoutPendingOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTimeoutPendingOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTimeoutPendingOrdersLogic {
	return &ListTimeoutPendingOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询存在超时待派单记录的订单 ID（去重），供 job 超时重派任务使用。
func (l *ListTimeoutPendingOrdersLogic) ListTimeoutPendingOrders(in *proto.ListTimeoutPendingOrdersRequest) (*proto.ListTimeoutPendingOrdersResponse, error) {
	timeoutSeconds := in.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultDispatchTimeoutSeconds
	}
	page := in.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > maxTimeoutOrderPageSize {
		pageSize = maxTimeoutOrderPageSize
	}

	before := time.Now().Add(-time.Duration(timeoutSeconds) * time.Second)
	orderIDs, total, err := l.svcCtx.DispatchRepository.ListTimeoutPendingOrderIDs(l.ctx, before, page, pageSize)
	if err != nil {
		return nil, err
	}

	resp := &proto.ListTimeoutPendingOrdersResponse{
		OrderIds: make([]int64, 0, len(orderIDs)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, id := range orderIDs {
		resp.OrderIds = append(resp.OrderIds, int64(id))
	}
	return resp, nil
}
