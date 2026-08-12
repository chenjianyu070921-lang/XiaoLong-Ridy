package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetVehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVehicleLogic {
	return &GetVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetVehicleLogic) GetVehicle(in *usersvc.GetVehicleReq) (*usersvc.GetVehicleResp, error) {
	// todo: add your logic here and delete this line

	return &usersvc.GetVehicleResp{}, nil
}
