package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// errEmptyCertification 表示未提供任何资质图片。
var errEmptyCertification = errors.New("至少上传一张资质图片")

// CertificationLogic 封装司机资质上传与查询逻辑，持有请求上下文与下游 driversvc 客户端。
type CertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCertificationLogic 构造司机资质逻辑处理器，注入请求上下文与服务上下文。
func NewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CertificationLogic {
	return &CertificationLogic{ctx: ctx, svcCtx: svcCtx}
}

// UploadCertification 上传司机资质图片：透传 base64 图片到 driversvc，由其直传 MinIO 并落库，返回资质记录。
// driverID 由鉴权中间件从 JWT 解析得到。
func (l *CertificationLogic) UploadCertification(driverID int64, req *types.UploadCertificationRequest) (*types.UploadCertificationResponse, error) {
	// 参数校验：司机ID和至少提供一张资质图片。
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	if req.IdCardFront == "" && req.IdCardBack == "" && req.DriverLicense == "" && req.VehicleLicense == "" {
		return nil, errEmptyCertification
	}
	if req.VehicleID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	vehicleResp, err := client.GetVehicle(l.ctx, &driversproto.GetVehicleRequest{Id: req.VehicleID})
	if err != nil {
		return nil, err
	}
	vehicle := vehicleResp.GetVehicle()
	if vehicle == nil {
		return nil, ErrInvalidParam
	}
	if vehicle.GetDriverId() != driverID {
		return nil, ErrForbiddenDriverResource
	}
	resp, err := client.UploadCertification(l.ctx, &driversproto.UploadCertificationRequest{
		DriverId:       driverID,
		VehicleId:      req.VehicleID,
		IdCardFront:    req.IdCardFront,
		IdCardBack:     req.IdCardBack,
		DriverLicense:  req.DriverLicense,
		VehicleLicense: req.VehicleLicense,
	})
	if err != nil {
		return nil, err
	}
	return &types.UploadCertificationResponse{
		ID:            resp.GetId(),
		Certification: toCertificationInfo(resp.GetCertification()),
	}, nil
}

// GetCertification 查询当前司机资质记录。
func (l *CertificationLogic) GetCertification(driverID int64) (*types.GetCertificationResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetCertification(l.ctx, &driversproto.GetCertificationRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	return &types.GetCertificationResponse{
		Certification: toCertificationInfo(resp.GetCertification()),
		Found:         resp.GetFound(),
	}, nil
}

// toCertificationInfo 将 proto 资质记录映射为 API 响应结构。
func toCertificationInfo(c *driversproto.CertificationInfo) *types.CertificationInfo {
	if c == nil {
		return nil
	}
	return &types.CertificationInfo{
		ID:                c.GetId(),
		DriverID:          c.GetDriverId(),
		VehicleID:         c.GetVehicleId(),
		IdCardFrontURL:    c.GetIdCardFrontUrl(),
		IdCardBackURL:     c.GetIdCardBackUrl(),
		DriverLicenseURL:  c.GetDriverLicenseUrl(),
		VehicleLicenseURL: c.GetVehicleLicenseUrl(),
		AuditStatus:       int(c.GetAuditStatus()),
		AuditRemark:       c.GetAuditRemark(),
	}
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *CertificationLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
