package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type FreezeUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFreezeUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeUserLogic {
	return &FreezeUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 冻结用户。
func (l *FreezeUserLogic) FreezeUser(in *adminsvc.ChangeUserStatusRequest) (*adminsvc.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.CommonResponse{}, nil
}
