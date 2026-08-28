package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

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

// GetVehicle 根据车辆 ID 查询车辆完整信息。
func (l *GetVehicleLogic) GetVehicle(in *proto.GetVehicleRequest) (*proto.GetVehicleResponse, error) {
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
	resp := &proto.GetVehicleResponse{
		Vehicle: &proto.Vehicle{
			Id:          int64(v.Id),
			DriverId:    int64(v.DriverId),
			PlateNo:     v.PlateNo,
			Brand:       v.Brand,
			Model:       v.Model,
			Color:       v.Color,
			VehicleType: int32(v.VehicleType),
			InsuranceNo: v.InsuranceNo,
			Status:      proto.VehicleStatus(v.Status),
			CreatedAt:   v.CreatedAt.Unix(),
			UpdatedAt:   v.UpdatedAt.Unix(),
		},
	}
	if v.RegistrationDate != nil {
		reg := v.RegistrationDate.Unix()
		resp.Vehicle.RegistrationDate = &reg
	}
	if v.InsuranceExpireAt != nil {
		exp := v.InsuranceExpireAt.Unix()
		resp.Vehicle.InsuranceExpireAt = &exp
	}
	return resp, nil
}
