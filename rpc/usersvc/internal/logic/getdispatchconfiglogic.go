package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDispatchConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDispatchConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDispatchConfigLogic {
	return &GetDispatchConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDispatchConfigLogic) GetDispatchConfig(in *usersvc.GetDispatchConfigReq) (*usersvc.GetDispatchConfigResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.GetDispatchConfigResp{}, nil
}
