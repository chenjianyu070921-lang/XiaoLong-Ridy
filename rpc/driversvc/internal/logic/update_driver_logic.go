package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDriverLogic {
	return &UpdateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDriverLogic) UpdateDriver(in *proto.UpdateDriverRequest) (*proto.UpdateDriverResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.UpdateDriverResponse{}, nil
}
