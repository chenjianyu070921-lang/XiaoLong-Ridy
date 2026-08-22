package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type VehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VehicleLogic {
	return &VehicleLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *VehicleLogic) CreateVehicle(driverID int64, req *types.CreateVehicleRequest) (*types.CreateVehicleResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	normalizeCreateVehicleRequest(req)
	if req.PlateNo == "" || req.Brand == "" || req.Model == "" || req.VehicleType <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateVehicle(l.ctx, &driversproto.CreateVehicleRequest{
		DriverId:          driverID,
		PlateNo:           req.PlateNo,
		Brand:             req.Brand,
		Model:             req.Model,
		Color:             req.Color,
		VehicleType:       req.VehicleType,
		RegistrationDate:  optionalPositiveInt64(req.RegistrationDate),
		InsuranceNo:       req.InsuranceNo,
		InsuranceExpireAt: optionalPositiveInt64(req.InsuranceExpireAt),
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateVehicleResponse{
		ID:        resp.GetId(),
		Status:    resp.GetStatus().String(),
		CreatedAt: resp.GetCreatedAt(),
	}, nil
}

func (l *VehicleLogic) GetVehicle(driverID, vehicleID int64) (*types.GetVehicleResponse, error) {
	if driverID <= 0 || vehicleID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetVehicle(l.ctx, &driversproto.GetVehicleRequest{Id: vehicleID})
	if err != nil {
		return nil, err
	}
	vehicle := resp.GetVehicle()
	if vehicle == nil {
		return nil, ErrInvalidParam
	}
	if vehicle.GetDriverId() != driverID {
		return nil, ErrForbiddenDriverResource
	}
	return &types.GetVehicleResponse{Vehicle: toVehicleInfo(vehicle)}, nil
}

func (l *VehicleLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

func normalizeCreateVehicleRequest(req *types.CreateVehicleRequest) {
	req.PlateNo = strings.ToUpper(strings.TrimSpace(req.PlateNo))
	req.Brand = strings.TrimSpace(req.Brand)
	req.Model = strings.TrimSpace(req.Model)
	req.Color = strings.TrimSpace(req.Color)
	req.InsuranceNo = strings.TrimSpace(req.InsuranceNo)
}

func optionalPositiveInt64(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func toVehicleInfo(v *driversproto.Vehicle) types.VehicleInfo {
	if v == nil {
		return types.VehicleInfo{}
	}
	return types.VehicleInfo{
		ID:                v.GetId(),
		DriverID:          v.GetDriverId(),
		PlateNo:           v.GetPlateNo(),
		Brand:             v.GetBrand(),
		Model:             v.GetModel(),
		Color:             v.GetColor(),
		VehicleType:       v.GetVehicleType(),
		RegistrationDate:  v.GetRegistrationDate(),
		InsuranceNo:       v.GetInsuranceNo(),
		InsuranceExpireAt: v.GetInsuranceExpireAt(),
		Status:            v.GetStatus().String(),
		CreatedAt:         v.GetCreatedAt(),
		UpdatedAt:         v.GetUpdatedAt(),
	}
}
