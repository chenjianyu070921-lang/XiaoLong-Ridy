// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDispatchSettingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDispatchSettingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDispatchSettingLogic {
	return &UpdateDispatchSettingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDispatchSettingLogic) UpdateDispatchSetting(req *types.DispatchSettingReq) (resp *types.DispatchSettingResp, err error) {
	// todo: add your logic here and delete this line

	return
}
