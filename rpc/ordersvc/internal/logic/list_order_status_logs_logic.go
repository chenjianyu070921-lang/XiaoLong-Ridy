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

func NewListOrderStatusLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrderStatusLogsLogic {
	return &ListOrderStatusLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListOrderStatusLogsLogic) ListOrderStatusLogs(in *proto.ListOrderStatusLogsRequest) (*proto.ListOrderStatusLogsResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.ListOrderStatusLogsResponse{}, nil
}
