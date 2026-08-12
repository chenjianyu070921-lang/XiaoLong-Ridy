package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type AcceptOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptOrderLogic {
	return &AcceptOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AcceptOrderLogic) AcceptOrder(in *proto.AcceptOrderRequest) (*proto.AcceptOrderResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.AcceptOrderResponse{}, nil
}
