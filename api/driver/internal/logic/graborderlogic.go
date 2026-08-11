// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GrabOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGrabOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GrabOrderLogic {
	return &GrabOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GrabOrderLogic) GrabOrder(req *types.GrabOrderReq) error {
	// todo: add your logic here and delete this line

	return nil
}
