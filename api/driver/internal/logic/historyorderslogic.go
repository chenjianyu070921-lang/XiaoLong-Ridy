// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HistoryOrdersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHistoryOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryOrdersLogic {
	return &HistoryOrdersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HistoryOrdersLogic) HistoryOrders(req *types.DriverIdOnly) (resp *types.OrderListResp, err error) {
	// todo: add your logic here and delete this line

	return
}
