package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateVehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVehicleLogic {
	return &UpdateVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateVehicle updates vehicle information by applying only explicitly set optional fields.
func (l *UpdateVehicleLogic) UpdateVehicle(in *proto.UpdateVehicleRequest) (*proto.UpdateVehicleResponse, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("vehicle id is invalid")
	}
	if l.svcCtx == nil || l.svcCtx.DriverVehicleRepository == nil {
		return nil, errors.New("driver vehicle repository not ready")
	}
	v, err := l.svcCtx.DriverVehicleRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if in.PlateNo != nil {
		updates["plate_no"] = in.GetPlateNo()
	}
	if in.Brand != nil {
		updates["brand"] = in.GetBrand()
	}
	if in.Model != nil {
		updates["model"] = in.GetModel()
	}
	if in.Color != nil {
		updates["color"] = in.GetColor()
	}
	if in.VehicleType != nil {
		updates["vehicle_type"] = int8(in.GetVehicleType())
	}
	if in.RegistrationDate != nil {
		t := time.Unix(in.GetRegistrationDate(), 0)
		updates["registration_date"] = &t
	}
	if in.InsuranceNo != nil {
		updates["insurance_no"] = in.GetInsuranceNo()
	}
	if in.InsuranceExpireAt != nil {
		t := time.Unix(in.GetInsuranceExpireAt(), 0)
		updates["insurance_expire_at"] = &t
	}
	if in.Status != nil {
		updates["status"] = int8(in.GetStatus())
	}

	if len(updates) == 0 {
		return nil, errors.New("no updatable fields")
	}
	if err := l.svcCtx.DriverVehicleRepository.Update(l.ctx, uint64(in.Id), updates); err != nil {
		return nil, err
	}
	if v, err = l.svcCtx.DriverVehicleRepository.GetByID(l.ctx, uint64(in.Id)); err != nil {
		return nil, err
	}
	return &proto.UpdateVehicleResponse{
		Id:        int64(v.Id),
		Status:    proto.VehicleStatus(v.Status),
		UpdatedAt: v.UpdatedAt.Unix(),
	}, nil
}
