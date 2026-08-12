package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitQualificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitQualificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitQualificationLogic {
	return &SubmitQualificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitQualificationLogic) SubmitQualification(in *usersvc.SubmitQualificationReq) (*usersvc.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.CommonResp{}, nil
}
