// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQualificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetQualificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQualificationLogic {
	return &GetQualificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQualificationLogic) GetQualification(req *types.DriverIdOnly) (resp *types.QualificationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
