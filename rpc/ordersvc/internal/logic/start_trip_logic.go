package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartTripLogic {
	return &StartTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StartTripLogic) StartTrip(in *proto.StartTripRequest) (*proto.StartTripResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.StartTripResponse{}, nil
}
