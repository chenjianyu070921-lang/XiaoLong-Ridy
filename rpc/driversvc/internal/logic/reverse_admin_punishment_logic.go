package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReverseAdminPunishmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReverseAdminPunishmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReverseAdminPunishmentLogic {
	return &ReverseAdminPunishmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReverseAdminPunishmentLogic) ReverseAdminPunishment(in *__proto.AdminPunishmentRequest) (*__proto.CommonResponse, error) {
	return applyAdminPunishment(l.ctx, l.svcCtx, in, true)
}
