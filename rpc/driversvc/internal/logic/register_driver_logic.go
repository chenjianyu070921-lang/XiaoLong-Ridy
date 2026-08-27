package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterDriverLogic {
	return &RegisterDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterDriverLogic) RegisterDriver(in *__proto.CreateDriverRequest) (*__proto.CreateDriverResponse, error) {
	return NewCreateDriverLogic(l.ctx, l.svcCtx).CreateDriver(in)
}
