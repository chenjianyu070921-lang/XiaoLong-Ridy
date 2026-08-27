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
	if !validCreateVehicle(req.PlateNo, req.Brand, req.Model, req.Color, req.VehicleType, req.InsuranceNo) {
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

func (l *VehicleLogic) UpdateVehicle(driverID int64, req *types.UpdateVehicleRequest) (*types.UpdateVehicleResponse, error) {
	if driverID <= 0 || req == nil || req.ID <= 0 {
		return nil, ErrInvalidParam
	}
	normalizeUpdateVehicleRequest(req)
	if !hasVehicleUpdateFields(req) {
		return nil, ErrInvalidParam
	}
	if req.PlateNo != nil && !validVehiclePlate(*req.PlateNo) {
		return nil, ErrInvalidParam
	}
	if req.Brand != nil && !validRequiredLength(*req.Brand, maxVehicleBrandLen) {
		return nil, ErrInvalidParam
	}
	if req.Model != nil && !validRequiredLength(*req.Model, maxVehicleModelLen) {
		return nil, ErrInvalidParam
	}
	if req.Color != nil && !validOptionalLength(*req.Color, maxVehicleColorLen) {
		return nil, ErrInvalidParam
	}
	if req.InsuranceNo != nil && !validOptionalLength(*req.InsuranceNo, maxVehicleInsuranceLen) {
		return nil, ErrInvalidParam
	}
	if req.VehicleType != nil && !validVehicleType(*req.VehicleType) {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	var driverIDPtr *int64
	if req.DriverID != nil {
		if *req.DriverID <= 0 {
			return nil, ErrInvalidParam
		}
		driverIDPtr = req.DriverID
	}
	resp, err := client.UpdateVehicle(l.ctx, &driversproto.UpdateVehicleRequest{
		Id:                req.ID,
		DriverId:          driverIDPtr,
		PlateNo:           req.PlateNo,
		Brand:             req.Brand,
		Model:             req.Model,
		Color:             req.Color,
		VehicleType:       req.VehicleType,
		RegistrationDate:  optionalPositiveInt64Ptr(req.RegistrationDate),
		InsuranceNo:       req.InsuranceNo,
		InsuranceExpireAt:  optionalPositiveInt64Ptr(req.InsuranceExpireAt),
		Status:            enumVehicleStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateVehicleResponse{
		ID:        resp.GetId(),
		Status:    resp.GetStatus().String(),
		UpdatedAt: resp.GetUpdatedAt(),
	}, nil
}

func (l *VehicleLogic) DeleteVehicle(driverID, vehicleID int64) (*types.DeleteVehicleResponse, error) {
	if driverID <= 0 || vehicleID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteVehicle(l.ctx, &driversproto.DeleteVehicleRequest{Id: vehicleID})
	if err != nil {
		return nil, err
	}
	return &types.DeleteVehicleResponse{
		ID:      resp.GetId(),
		Success: resp.GetSuccess(),
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

func normalizeUpdateVehicleRequest(req *types.UpdateVehicleRequest) {
	if req == nil {
		return
	}
	if req.PlateNo != nil {
		v := strings.ToUpper(strings.TrimSpace(*req.PlateNo))
		req.PlateNo = &v
	}
	if req.Brand != nil {
		v := strings.TrimSpace(*req.Brand)
		req.Brand = &v
	}
	if req.Model != nil {
		v := strings.TrimSpace(*req.Model)
		req.Model = &v
	}
	if req.Color != nil {
		v := strings.TrimSpace(*req.Color)
		req.Color = &v
	}
	if req.InsuranceNo != nil {
		v := strings.TrimSpace(*req.InsuranceNo)
		req.InsuranceNo = &v
	}
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		req.Status = &v
	}
}

func hasVehicleUpdateFields(req *types.UpdateVehicleRequest) bool {
	return req.DriverID != nil ||
		req.PlateNo != nil ||
		req.Brand != nil ||
		req.Model != nil ||
		req.Color != nil ||
		req.VehicleType != nil ||
		req.RegistrationDate != nil ||
		req.InsuranceNo != nil ||
		req.InsuranceExpireAt != nil ||
		req.Status != nil
}

func optionalPositiveInt64(value int64) *int64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func optionalPositiveInt64Ptr(value *int64) *int64 {
	if value == nil || *value <= 0 {
		return nil
	}
	v := *value
	return &v
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
