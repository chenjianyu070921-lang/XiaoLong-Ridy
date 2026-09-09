package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVehiclesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVehiclesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVehiclesLogic {
	return &ListVehiclesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListVehicles 按司机 ID 列出其绑定的全部车辆，支撑司机端最多绑定 3 辆、展示 3 辆的车辆管理。
func (l *ListVehiclesLogic) ListVehicles(in *__proto.ListVehiclesRequest) (*__proto.ListVehiclesResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if l.svcCtx == nil || l.svcCtx.DriverVehicleRepository == nil {
		return nil, errors.New("driver vehicle repository not ready")
	}
	vehicles, err := l.svcCtx.DriverVehicleRepository.ListByDriverID(l.ctx, uint64(in.DriverId))
	if err != nil {
		return nil, err
	}
	resp := &__proto.ListVehiclesResponse{}
	for _, v := range vehicles {
		pv := &__proto.Vehicle{
			Id:          int64(v.Id),
			DriverId:    int64(v.DriverId),
			PlateNo:     v.PlateNo,
			Brand:       v.Brand,
			Model:       v.Model,
			Color:       v.Color,
			VehicleType: int32(v.VehicleType),
			InsuranceNo: v.InsuranceNo,
			Status:      __proto.VehicleStatus(v.Status),
			CreatedAt:   v.CreatedAt.Unix(),
			UpdatedAt:   v.UpdatedAt.Unix(),
		}
		if v.RegistrationDate != nil {
			reg := v.RegistrationDate.Unix()
			pv.RegistrationDate = &reg
		}
		if v.InsuranceExpireAt != nil {
			exp := v.InsuranceExpireAt.Unix()
			pv.InsuranceExpireAt = &exp
		}
		resp.Vehicles = append(resp.Vehicles, pv)
	}
	return resp, nil
}
