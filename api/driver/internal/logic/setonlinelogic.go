// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetOnlineLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetOnlineLogic {
	return &SetOnlineLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetOnlineLogic) SetOnline(req *types.OnlineReq) error {
	// todo: add your logic here and delete this line

	return nil
}
