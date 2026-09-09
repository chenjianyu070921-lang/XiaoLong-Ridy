package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginBySmsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginBySmsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginBySmsLogic {
	return &LoginBySmsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginBySmsLogic) LoginBySms(in *__proto.LoginBySMSRequest) (*__proto.LoginResponse, error) {
	return NewLoginLogic(l.ctx, l.svcCtx).LoginBySMS(in)
}
