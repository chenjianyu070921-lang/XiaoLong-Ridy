package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDispatchConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDispatchConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDispatchConfigLogic {
	return &UpdateDispatchConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDispatchConfigLogic) UpdateDispatchConfig(in *usersvc.UpdateDispatchConfigReq) (*usersvc.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.CommonResp{}, nil
}
