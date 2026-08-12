package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"
	"XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type FinishTripLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFinishTripLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FinishTripLogic {
	return &FinishTripLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FinishTripLogic) FinishTrip(in *proto.FinishTripRequest) (*proto.FinishTripResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.FinishTripResponse{}, nil
}
