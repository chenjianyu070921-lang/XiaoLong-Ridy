package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQualificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetQualificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQualificationLogic {
	return &GetQualificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetQualificationLogic) GetQualification(in *usersvc.GetQualificationReq) (*usersvc.GetQualificationResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.GetQualificationResp{}, nil
}
