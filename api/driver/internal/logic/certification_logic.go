package logic

import (
	"context"
	"regexp"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// CertificationLogic 封装司机资质上传与查询逻辑，持有请求上下文与下游 driversvc 客户端。
type CertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCertificationLogic 构造司机资质逻辑处理器，注入请求上下文与服务上下文。
func NewCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CertificationLogic {
	return &CertificationLogic{ctx: ctx, svcCtx: svcCtx}
}

// UploadCertification 提交司机资质审核：以身份证号、真实姓名、驾驶证编号进行资质校验，
// 不再接收资质图片（图片上传链路已废弃，proto 字段与 MinIO 存储代码保留但不再使用）。
// driverID 由鉴权中间件从 JWT 解析得到。
func (l *CertificationLogic) UploadCertification(driverID int64, req *types.UploadCertificationRequest) (*types.UploadCertificationResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	if !isValidIdCardNo(req.IdCardNo) || !isValidRealName(req.RealName) || !isValidDriverLicenseNo(req.DriverLicenseNo) {
		return nil, ErrInvalidParam
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
	// 图片字段故意留空：资质图片上传已废弃，资质校验改用身份证号/姓名/驾照编号，
	// driversvc 的 CertificationInfo 图片 URL 因此保持为空，MinIO 直传不会被触发。
	resp, err := client.UploadCertification(l.ctx, &driversproto.UploadCertificationRequest{
		DriverId:  driverID,
		VehicleId: req.VehicleID,
	})
	if err != nil {
		return nil, err
	}
	return &types.UploadCertificationResponse{
		ID:            resp.GetId(),
		Certification: toCertificationInfo(resp.GetCertification()),
	}, nil
}

var idCardNoPattern = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

// isValidIdCardNo 校验 18 位居民身份证号码格式（地区+出生日期+顺序+校验位，末位可为 X）。
func isValidIdCardNo(value string) bool {
	return idCardNoPattern.MatchString(strings.TrimSpace(value))
}

// isValidRealName 校验真实姓名非空且不少于 2 个字符。
func isValidRealName(value string) bool {
	name := strings.TrimSpace(value)
	return name != "" && len([]rune(name)) >= 2
}

// isValidDriverLicenseNo 校验驾驶证编号非空。
func isValidDriverLicenseNo(value string) bool {
	return strings.TrimSpace(value) != ""
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
