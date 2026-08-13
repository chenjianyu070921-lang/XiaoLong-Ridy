// Package logic 实现 driver API 的业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"   // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"  // API 层使用的请求/响应类型
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// VehicleLogic 车辆业务逻辑处理器，持有上下文与下游客户端。
type VehicleLogic struct {
	ctx    context.Context    // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewVehicleLogic 构造车辆逻辑处理器实例。
func NewVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VehicleLogic {
	// 注入上下文与服务上下文。
	return &VehicleLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateVehicle 创建车辆，校验司机归属、车牌、品牌与车辆类型。
func (l *VehicleLogic) CreateVehicle(req *types.CreateVehicleRequest) (*types.CreateVehicleResponse, error) {
	// 校验所属司机 ID 合法性。
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 校验车牌号非空。
	if req.PlateNo == "" {
		return nil, errors.New("车牌号不能为空")
	}
	// 校验品牌非空。
	if req.Brand == "" {
		return nil, errors.New("品牌不能为空")
	}
	// 校验车辆类型在合法枚举范围 1~3 内。
	if req.VehicleType < 1 || req.VehicleType > 3 {
		return nil, errors.New("车辆类型不合法(1特惠快车 2快车 3拼车)")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游创建车辆接口，映射 API 入参为 proto 请求。
	resp, err := client.CreateVehicle(l.ctx, &driversproto.CreateVehicleRequest{
		DriverId:          req.DriverID,          // 所属司机 ID
		PlateNo:           req.PlateNo,           // 车牌号
		Brand:             req.Brand,             // 品牌
		Model:             req.Model,             // 车型
		Color:             req.Color,             // 颜色
		VehicleType:       req.VehicleType,       // 车辆类型
		RegistrationDate:  req.RegistrationDate,  // 注册日期
		InsuranceNo:       req.InsuranceNo,       // 保险单号
		InsuranceExpireAt: req.InsuranceExpireAt, // 保险到期日
	})
	if err != nil {
		return nil, err
	}
	// 返回创建结果（车辆 ID + 初始状态）。
	return &types.CreateVehicleResponse{ID: resp.GetId(), Status: resp.GetStatus()}, nil
}

// UpdateVehicle 更新车辆，校验 ID 与可选车辆类型范围。
func (l *VehicleLogic) UpdateVehicle(req *types.UpdateVehicleRequest) (*types.UpdateVehicleResponse, error) {
	// 校验车辆 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	// 若传入车辆类型，校验其范围 1~3。
	if req.VehicleType != nil && (*req.VehicleType < 1 || *req.VehicleType > 3) {
		return nil, errors.New("车辆类型不合法(1特惠快车 2快车 3拼车)")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口，可选字段直接透传指针。
	resp, err := client.UpdateVehicle(l.ctx, &driversproto.UpdateVehicleRequest{
		Id:                req.ID,                // 车辆 ID
		DriverId:          req.DriverID,          // 可选司机 ID
		PlateNo:           req.PlateNo,           // 可选车牌号
		Brand:             req.Brand,             // 可选品牌
		Model:             req.Model,             // 可选车型
		Color:             req.Color,             // 可选颜色
		VehicleType:       req.VehicleType,       // 可选车辆类型
		RegistrationDate:  req.RegistrationDate,  // 可选注册日期
		InsuranceNo:       req.InsuranceNo,       // 可选保险单号
		InsuranceExpireAt: req.InsuranceExpireAt, // 可选保险到期日
		Status:            req.Status,            // 可选状态
	})
	if err != nil {
		return nil, err
	}
	// 返回更新结果。
	return &types.UpdateVehicleResponse{ID: resp.GetId(), Status: resp.GetStatus(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetVehicle 查询车辆详情。
func (l *VehicleLogic) GetVehicle(id int64) (*types.GetVehicleResponse, error) {
	// 校验车辆 ID 合法性。
	if id <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetVehicle(l.ctx, &driversproto.GetVehicleRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的车辆实体。
	v := resp.GetVehicle()
	// 映射为 API 的车辆详情结构并返回。
	return &types.GetVehicleResponse{Vehicle: types.VehicleDetail{
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
		Status:            v.GetStatus(),
		CreatedAt:         v.GetCreatedAt(),
		UpdatedAt:         v.GetUpdatedAt(),
	}}, nil
}

// DeleteVehicle 删除车辆。
func (l *VehicleLogic) DeleteVehicle(id int64) (*types.DeleteResponse, error) {
	// 校验车辆 ID 合法性。
	if id <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游删除接口。
	resp, err := client.DeleteVehicle(l.ctx, &driversproto.DeleteVehicleRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 返回删除结果。
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// ListVehicles 分页查询车辆列表。
func (l *VehicleLogic) ListVehicles(req *types.ListVehiclesRequest) (*types.ListVehiclesResponse, error) {
	// 收敛分页参数到合法范围。
	page, pageSize := clampPage(req.Page, req.PageSize)
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游列表接口。
	resp, err := client.ListVehicles(l.ctx, &driversproto.ListVehiclesRequest{
		DriverId: req.DriverID, Status: req.Status, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	// 预分配切片。
	list := make([]types.VehicleSummary, 0, len(resp.GetList()))
	// 遍历并映射为 API 摘要结构。
	for _, s := range resp.GetList() {
		list = append(list, types.VehicleSummary{
			ID:          s.GetId(),
			DriverID:    s.GetDriverId(),
			PlateNo:     s.GetPlateNo(),
			Brand:       s.GetBrand(),
			VehicleType: s.GetVehicleType(),
			Status:      s.GetStatus(),
		})
	}
	// 组装分页响应返回。
	return &types.ListVehiclesResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// client 从服务上下文中安全取出 driversvc 客户端。
func (l *VehicleLogic) client() (svc.DriverClient, error) {
	// 防御性校验客户端是否可用。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
