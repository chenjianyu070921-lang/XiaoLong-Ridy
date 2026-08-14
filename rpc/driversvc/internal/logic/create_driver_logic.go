package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDriverLogic {
	return &CreateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDriverLogic) CreateDriver(in *proto.CreateDriverRequest) (*proto.CreateDriverResponse, error) {
	// todo: add your logic here and delete this line

	return &proto.CreateDriverResponse{}, nil
}
