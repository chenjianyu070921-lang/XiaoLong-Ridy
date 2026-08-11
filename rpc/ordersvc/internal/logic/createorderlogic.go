package logic

import (
	"XiaoLong-Ridy/rpc/ordersvc/proto"
	"context"

	"XiaoLong-Ridy/rpc/ordersvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOrderLogic) CreateOrder(in *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.CreateOrderResponse{
		OrderId:             0,
		OrderNo:             "",
		EstimatedPriceCents: 0,
		Status:              0,
		CreatedAt:           0,
	}, nil
}
