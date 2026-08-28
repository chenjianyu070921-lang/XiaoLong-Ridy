package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteVehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVehicleLogic {
	return &DeleteVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteVehicle 根据车辆 ID 删除车辆（driver_vehicle 表无软删字段，按物理删除处理）。
func (l *DeleteVehicleLogic) DeleteVehicle(in *proto.DeleteVehicleRequest) (*proto.DeleteVehicleResponse, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	if l.svcCtx == nil || l.svcCtx.DriverVehicleRepository == nil {
		return nil, errors.New("driver vehicle repository not ready")
	}
	v, err := l.svcCtx.DriverVehicleRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.DriverVehicleRepository.Delete(l.ctx, v); err != nil {
		return nil, err
	}
	return &proto.DeleteVehicleResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
