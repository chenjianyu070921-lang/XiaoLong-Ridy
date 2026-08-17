package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAbnormalOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAbnormalOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAbnormalOrdersLogic {
	return &ListAbnormalOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询异常订单列表。
func (l *ListAbnormalOrdersLogic) ListAbnormalOrders(in *adminsvc.AbnormalOrderListRequest) (*adminsvc.AbnormalOrderListResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.AbnormalOrderListResponse{}, nil
}
