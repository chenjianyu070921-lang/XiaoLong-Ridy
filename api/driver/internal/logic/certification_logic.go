// Package logic 实现 driver API 的业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"   // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"  // API 层使用的请求/响应类型
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// CertificationLogic 认证业务逻辑处理器，持有上下文与下游客户端。
type CertificationLogic struct {
	ctx    context.Context    // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewCertificationLogic 构造认证逻辑处理器实例。
func NewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CertificationLogic {
	// 注入上下文与服务上下文。
	return &CertificationLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateCertification 创建认证，校验司机/车辆归属与四张图片地址。
func (l *CertificationLogic) CreateCertification(req *types.CreateCertificationRequest) (*types.CreateCertificationResponse, error) {
	// 校验所属司机 ID 合法性。
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 校验关联车辆 ID 合法性。
	if req.VehicleID <= 0 {
		return nil, errors.New("车辆ID不合法")
	}
	// 四张认证图片地址任一为空即不合法。
	if req.IdCardFrontURL == "" || req.IdCardBackURL == "" || req.DriverLicenseURL == "" || req.VehicleLicenseURL == "" {
		return nil, errors.New("四张认证图片地址均不能为空")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游创建认证接口。
	resp, err := client.CreateCertification(l.ctx, &driversproto.CreateCertificationRequest{
		DriverId:          req.DriverID,          // 司机 ID
		VehicleId:         req.VehicleID,         // 车辆 ID
		IdCardFrontUrl:    req.IdCardFrontURL,    // 身份证人像面
		IdCardBackUrl:     req.IdCardBackURL,     // 身份证国徽面
		DriverLicenseUrl:  req.DriverLicenseURL,  // 驾驶证照片
		VehicleLicenseUrl: req.VehicleLicenseURL, // 行驶证照片
	})
	if err != nil {
		return nil, err
	}
	// 返回创建结果（认证 ID + 初始审核状态）。
	return &types.CreateCertificationResponse{ID: resp.GetId(), AuditStatus: resp.GetAuditStatus()}, nil
}

// UpdateCertification 更新认证，校验 ID 与可选审核状态范围。
func (l *CertificationLogic) UpdateCertification(req *types.UpdateCertificationRequest) (*types.UpdateCertificationResponse, error) {
	// 校验认证 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	// 若传入审核状态，校验其范围 1~3。
	if req.AuditStatus != nil && (*req.AuditStatus < 1 || *req.AuditStatus > 3) {
		return nil, errors.New("审核状态不合法(1待审核 2通过 3驳回)")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口，可选字段直接透传指针。
	resp, err := client.UpdateCertification(l.ctx, &driversproto.UpdateCertificationRequest{
		Id:                req.ID,                // 认证 ID
		DriverId:          req.DriverID,          // 可选司机 ID
		VehicleId:         req.VehicleID,         // 可选车辆 ID
		IdCardFrontUrl:    req.IdCardFrontURL,    // 可选身份证人像面
		IdCardBackUrl:     req.IdCardBackURL,     // 可选身份证国徽面
		DriverLicenseUrl:  req.DriverLicenseURL,  // 可选驾驶证照片
		VehicleLicenseUrl: req.VehicleLicenseURL, // 可选行驶证照片
		AuditStatus:       req.AuditStatus,       // 可选审核状态
		AuditRemark:       req.AuditRemark,       // 可选审核备注
		AuditedBy:         req.AuditedBy,         // 可选审核人
		AuditedAt:         req.AuditedAt,         // 可选审核时间
	})
	if err != nil {
		return nil, err
	}
	// 返回更新结果。
	return &types.UpdateCertificationResponse{ID: resp.GetId(), AuditStatus: resp.GetAuditStatus(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetCertification 查询认证详情。
func (l *CertificationLogic) GetCertification(id int64) (*types.GetCertificationResponse, error) {
	// 校验认证 ID 合法性。
	if id <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetCertification(l.ctx, &driversproto.GetCertificationRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的认证实体。
	c := resp.GetCertification()
	// 映射为 API 的认证详情结构并返回。
	return &types.GetCertificationResponse{Certification: types.CertificationDetail{
		ID:                c.GetId(),
		DriverID:          c.GetDriverId(),
		VehicleID:         c.GetVehicleId(),
		IdCardFrontURL:    c.GetIdCardFrontUrl(),
		IdCardBackURL:     c.GetIdCardBackUrl(),
		DriverLicenseURL:  c.GetDriverLicenseUrl(),
		VehicleLicenseURL: c.GetVehicleLicenseUrl(),
		AuditStatus:       c.GetAuditStatus(),
		AuditRemark:       c.GetAuditRemark(),
		AuditedBy:         c.GetAuditedBy(),
		AuditedAt:         c.GetAuditedAt(),
		CreatedAt:         c.GetCreatedAt(),
		UpdatedAt:         c.GetUpdatedAt(),
	}}, nil
}

// DeleteCertification 删除认证。
func (l *CertificationLogic) DeleteCertification(id int64) (*types.DeleteResponse, error) {
	// 校验认证 ID 合法性。
	if id <= 0 {
		return nil, errors.New("认证ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游删除接口。
	resp, err := client.DeleteCertification(l.ctx, &driversproto.DeleteCertificationRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 返回删除结果。
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// ListCertifications 分页查询认证列表。
func (l *CertificationLogic) ListCertifications(req *types.ListCertificationsRequest) (*types.ListCertificationsResponse, error) {
	// 收敛分页参数到合法范围。
	page, pageSize := clampPage(req.Page, req.PageSize)
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游列表接口。
	resp, err := client.ListCertifications(l.ctx, &driversproto.ListCertificationsRequest{
		DriverId: req.DriverID, AuditStatus: req.AuditStatus, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	// 预分配切片。
	list := make([]types.CertificationSummary, 0, len(resp.GetList()))
	// 遍历并映射为 API 摘要结构。
	for _, s := range resp.GetList() {
		list = append(list, types.CertificationSummary{
			ID:          s.GetId(),
			DriverID:    s.GetDriverId(),
			VehicleID:   s.GetVehicleId(),
			AuditStatus: s.GetAuditStatus(),
			CreatedAt:   s.GetCreatedAt(),
		})
	}
	// 组装分页响应返回。
	return &types.ListCertificationsResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// client 从服务上下文中安全取出 driversvc 客户端。
func (l *CertificationLogic) client() (svc.DriverClient, error) {
	// 防御性校验客户端是否可用。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
