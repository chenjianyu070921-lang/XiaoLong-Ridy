package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmArriveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmArriveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmArriveLogic {
	return &ConfirmArriveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmArriveLogic) ConfirmArrive(in *proto.ConfirmArriveRequest) (*proto.ConfirmArriveResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.ConfirmArriveResponse{}, nil
}
