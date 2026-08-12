package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBankCardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBankCardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBankCardLogic {
	return &UpdateBankCardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateBankCardLogic) UpdateBankCard(in *usersvc.UpdateBankCardReq) (*usersvc.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.CommonResp{}, nil
}
