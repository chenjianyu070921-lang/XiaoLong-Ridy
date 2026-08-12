package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimeoutCancelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTimeoutCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TimeoutCancelLogic {
	return &TimeoutCancelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *TimeoutCancelLogic) TimeoutCancel(in *proto.TimeoutCancelRequest) (*proto.TimeoutCancelResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.TimeoutCancelResponse{}, nil
}
