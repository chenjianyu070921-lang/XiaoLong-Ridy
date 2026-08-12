package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitVehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitVehicleLogic {
	return &SubmitVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SubmitVehicleLogic) SubmitVehicle(in *usersvc.SubmitVehicleReq) (*usersvc.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.CommonResp{}, nil
}
