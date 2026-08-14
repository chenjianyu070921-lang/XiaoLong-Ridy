package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDriverLogic {
	return &DeleteDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteDriverLogic) DeleteDriver(in *proto.DeleteDriverRequest) (*proto.DeleteDriverResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.DeleteDriverResponse{}, nil
}
