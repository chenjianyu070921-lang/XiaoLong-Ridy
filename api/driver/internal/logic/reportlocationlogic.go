// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"driver/internal/svc"
	"driver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportLocationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReportLocationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportLocationLogic {
	return &ReportLocationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReportLocationLogic) ReportLocation(req *types.LocationReportReq) error {
	// todo: add your logic here and delete this line

	return nil
}
