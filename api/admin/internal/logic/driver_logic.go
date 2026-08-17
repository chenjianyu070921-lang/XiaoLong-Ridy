package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// DriverLogic 负责管理后台司机审核相关动作的 HTTP 适配。
type DriverLogic struct {
	ctx *svc.ServiceContext
}

// NewDriverLogic 创建司机审核逻辑。
func NewDriverLogic(ctx *svc.ServiceContext) *DriverLogic {
	return &DriverLogic{ctx: ctx}
}

// ListCertifications 查询司机审核列表。
func (l *DriverLogic) ListCertifications(ctx context.Context, req types.DriverCertificationListRequest) (*types.PageResult, error) {
	resp, err := l.ctx.AdminSvc.ListDriverCertifications(ctx, &adminclient.DriverCertificationListRequest{
		Page:        int32(req.Page),
		PageSize:    int32(req.PageSize),
		Keyword:     req.Keyword,
		AuditStatus: req.AuditStatus,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverCertificationDTO, 0, len(resp.List))
	for _, item := range resp.List {
		items = append(items, types.DriverCertificationDTO{
			ID:                item.Id,
			DriverID:          item.DriverId,
			VehicleID:         item.VehicleId,
			DriverPhone:       item.DriverPhone,
			DriverName:        item.DriverName,
			DriverStatus:      item.DriverStatus,
			PlateNo:           item.PlateNo,
			VehicleStatus:     item.VehicleStatus,
			IDCardFrontURL:    item.IdCardFrontUrl,
			IDCardBackURL:     item.IdCardBackUrl,
			DriverLicenseURL:  item.DriverLicenseUrl,
			VehicleLicenseURL: item.VehicleLicenseUrl,
			AuditStatus:       item.AuditStatus,
			AuditRemark:       item.AuditRemark,
			AuditedBy:         item.AuditedBy,
			AuditedAt:         item.AuditedAt,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt,
		})
	}
	return &types.PageResult{List: items, Total: resp.Total, Page: int(resp.Page), PageSize: int(resp.PageSize)}, nil
}

// CertificationDetail 查询司机审核详情。
func (l *DriverLogic) CertificationDetail(ctx context.Context, id int64) (*types.DriverCertificationDTO, error) {
	resp, err := l.ctx.AdminSvc.GetDriverCertification(ctx, &adminclient.DriverCertificationDetailRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DriverCertificationDTO{
		ID:                resp.Id,
		DriverID:          resp.DriverId,
		VehicleID:         resp.VehicleId,
		DriverPhone:       resp.DriverPhone,
		DriverName:        resp.DriverName,
		DriverStatus:      resp.DriverStatus,
		PlateNo:           resp.PlateNo,
		VehicleStatus:     resp.VehicleStatus,
		IDCardFrontURL:    resp.IdCardFrontUrl,
		IDCardBackURL:     resp.IdCardBackUrl,
		DriverLicenseURL:  resp.DriverLicenseUrl,
		VehicleLicenseURL: resp.VehicleLicenseUrl,
		AuditStatus:       resp.AuditStatus,
		AuditRemark:       resp.AuditRemark,
		AuditedBy:         resp.AuditedBy,
		AuditedAt:         resp.AuditedAt,
		CreatedAt:         resp.CreatedAt,
		UpdatedAt:         resp.UpdatedAt,
	}, nil
}

// ApproveCertification 审核通过司机认证。
func (l *DriverLogic) ApproveCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "审核通过"
	}
	_, err := l.ctx.AdminSvc.ApproveDriverCertification(ctx, &adminclient.AuditDriverCertificationRequest{
		Id:      id,
		Remark:  remark,
		AdminId: session.AdminID,
		Ip:      ip,
	})
	return err
}

// RejectCertification 驳回司机认证。
func (l *DriverLogic) RejectCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "资料不完整"
	}
	_, err := l.ctx.AdminSvc.RejectDriverCertification(ctx, &adminclient.AuditDriverCertificationRequest{
		Id:      id,
		Remark:  remark,
		AdminId: session.AdminID,
		Ip:      ip,
	})
	return err
}
