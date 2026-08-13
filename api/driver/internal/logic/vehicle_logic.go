package logic

import (
	"context"
	"errors"

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

func (l *VehicleLogic) CreateVehicle(req *types.CreateVehicleRequest) (*types.CreateVehicleResponse, error) {
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if req.PlateNo == "" {
		return nil, errors.New("车牌号不能为空")
	}
	if req.Brand == "" {
		return nil, errors.New("品牌不能为空")
	}
	if req.VehicleType < 1 || req.VehicleType > 3 {
		return nil, errors.New("车辆类型不合法(1特惠快车 2快车 3拼车)")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateVehicle(l.ctx, &driversproto.CreateVehicleRequest{
		DriverId:          req.DriverID,
		PlateNo:           req.PlateNo,
		Brand:             req.Brand,
		Model:             req.Model,
		Color:             req.Color,
		VehicleType:       req.VehicleType,
		RegistrationDate:  req.RegistrationDate,
		InsuranceNo:       req.InsuranceNo,
		InsuranceExpireAt: req.InsuranceExpireAt,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateVehicleResponse{ID: resp.GetId(), Status: resp.GetStatus()}, nil
}

func (l *VehicleLogic) UpdateVehicle(req *types.UpdateVehicleRequest) (*types.UpdateVehicleResponse, error) {
	if req.ID <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	if req.VehicleType != nil && (*req.VehicleType < 1 || *req.VehicleType > 3) {
		return nil, errors.New("车辆类型不合法(1特惠快车 2快车 3拼车)")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateVehicle(l.ctx, &driversproto.UpdateVehicleRequest{
		Id:                req.ID,
		DriverId:          req.DriverID,
		PlateNo:           req.PlateNo,
		Brand:             req.Brand,
		Model:             req.Model,
		Color:             req.Color,
		VehicleType:       req.VehicleType,
		RegistrationDate:  req.RegistrationDate,
		InsuranceNo:       req.InsuranceNo,
		InsuranceExpireAt: req.InsuranceExpireAt,
		Status:            req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateVehicleResponse{ID: resp.GetId(), Status: resp.GetStatus(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

func (l *VehicleLogic) GetVehicle(id int64) (*types.GetVehicleResponse, error) {
	if id <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetVehicle(l.ctx, &driversproto.GetVehicleRequest{Id: id})
	if err != nil {
		return nil, err
	}
	v := resp.GetVehicle()
	return &types.GetVehicleResponse{Vehicle: types.VehicleDetail{
		ID: v.GetId(), DriverID: v.GetDriverId(), PlateNo: v.GetPlateNo(), Brand: v.GetBrand(), Model: v.GetModel(),
		Color: v.GetColor(), VehicleType: v.GetVehicleType(), RegistrationDate: v.GetRegistrationDate(),
		InsuranceNo: v.GetInsuranceNo(), InsuranceExpireAt: v.GetInsuranceExpireAt(), Status: v.GetStatus(),
		CreatedAt: v.GetCreatedAt(), UpdatedAt: v.GetUpdatedAt(),
	}}, nil
}

func (l *VehicleLogic) DeleteVehicle(id int64) (*types.DeleteResponse, error) {
	if id <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteVehicle(l.ctx, &driversproto.DeleteVehicleRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

func (l *VehicleLogic) ListVehicles(req *types.ListVehiclesRequest) (*types.ListVehiclesResponse, error) {
	page, pageSize := clampPage(req.Page, req.PageSize)
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListVehicles(l.ctx, &driversproto.ListVehiclesRequest{
		DriverId: req.DriverID, Status: req.Status, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.VehicleSummary, 0, len(resp.GetList()))
	for _, s := range resp.GetList() {
		list = append(list, types.VehicleSummary{
			ID: s.GetId(), DriverID: s.GetDriverId(), PlateNo: s.GetPlateNo(),
			Brand: s.GetBrand(), VehicleType: s.GetVehicleType(), Status: s.GetStatus(),
		})
	}
	return &types.ListVehiclesResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

func (l *VehicleLogic) client() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
