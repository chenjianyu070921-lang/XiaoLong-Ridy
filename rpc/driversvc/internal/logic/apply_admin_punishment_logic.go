package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyAdminPunishmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApplyAdminPunishmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyAdminPunishmentLogic {
	return &ApplyAdminPunishmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApplyAdminPunishmentLogic) ApplyAdminPunishment(in *__proto.AdminPunishmentRequest) (*__proto.CommonResponse, error) {
	return applyAdminPunishment(l.ctx, l.svcCtx, in, false)
}
