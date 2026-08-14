package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLogic {
	return &GetDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDriverLogic) GetDriver(in *proto.GetDriverRequest) (*proto.GetDriverResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.GetDriverResponse{}, nil
}
