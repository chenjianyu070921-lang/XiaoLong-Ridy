package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
)

// DriverLogic 封装管理后台司机审核业务。
// 它负责审核状态校验、调用仓储更新数据，以及写入后台操作日志。
type DriverLogic struct {
	ctx *svc.ServiceContext
}

// NewDriverLogic 创建司机业务逻辑对象。
func NewDriverLogic(ctx *svc.ServiceContext) *DriverLogic {
	return &DriverLogic{ctx: ctx}
}

// ListCertifications 查询司机认证审核列表。
func (l *DriverLogic) ListCertifications(ctx context.Context, req types.DriverCertificationListRequest) (*types.PageResult, error) {
	list, total, err := l.ctx.DriverRepository.ListCertifications(ctx, req)
	if err != nil {
		return nil, err
	}
	items := make([]types.DriverCertificationDTO, 0, len(list))
	for _, item := range list {
		items = append(items, toCertificationDTO(item))
	}
	return &types.PageResult{
		List:     items,
		Total:    total,
		Page:     normalizePage(req.Page),
		PageSize: normalizePageSize(req.PageSize),
	}, nil
}

// CertificationDetail 查询司机认证详情。
func (l *DriverLogic) CertificationDetail(ctx context.Context, id int64) (*types.DriverCertificationDTO, error) {
	item, err := l.ctx.DriverRepository.GetCertificationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toCertificationDTO(*item)
	return &dto, nil
}

// ApproveCertification 审核通过司机认证。
// 通过后会同步更新司机账号和车辆状态，并记录 admin_operation_log。
func (l *DriverLogic) ApproveCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		remark = "审核通过"
	}
	if err := l.ctx.DriverRepository.AuditCertification(ctx, id, 2, remark, session.AdminID); err != nil {
		return err
	}
	return l.ctx.OperationLogRepository.Create(ctx, repository.CreateOperationLogInput{
		AdminID:    session.AdminID,
		Module:     "driver",
		Action:     "approve_certification",
		TargetType: "driver_certification",
		TargetID:   id,
		Detail:     remark,
		IP:         ip,
	})
}

// RejectCertification 驳回司机认证。
// 驳回必须填写原因，方便司机端展示补充资料要求。
func (l *DriverLogic) RejectCertification(ctx context.Context, id int64, req types.AuditRequest, session *model.AdminSession, ip string) error {
	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		return ErrBadRequest
	}
	if err := l.ctx.DriverRepository.AuditCertification(ctx, id, 3, remark, session.AdminID); err != nil {
		return err
	}
	return l.ctx.OperationLogRepository.Create(ctx, repository.CreateOperationLogInput{
		AdminID:    session.AdminID,
		Module:     "driver",
		Action:     "reject_certification",
		TargetType: "driver_certification",
		TargetID:   id,
		Detail:     remark,
		IP:         ip,
	})
}

// toCertificationDTO 将司机认证聚合模型转换为接口 DTO。
func toCertificationDTO(item model.DriverCertification) types.DriverCertificationDTO {
	return types.DriverCertificationDTO{
		ID:                item.ID,
		DriverID:          item.DriverID,
		VehicleID:         item.VehicleID,
		DriverPhone:       item.DriverPhone,
		DriverName:        item.DriverName,
		DriverStatus:      item.DriverStatus,
		PlateNo:           item.PlateNo,
		VehicleStatus:     item.VehicleStatus,
		IDCardFrontURL:    item.IDCardFrontURL,
		IDCardBackURL:     item.IDCardBackURL,
		DriverLicenseURL:  item.DriverLicenseURL,
		VehicleLicenseURL: item.VehicleLicenseURL,
		AuditStatus:       item.AuditStatus,
		AuditRemark:       item.AuditRemark,
		AuditedBy:         item.AuditedBy,
		AuditedAt:         repository.FormatOptionalTime(item.AuditedAt),
		CreatedAt:         repository.FormatTime(item.CreatedAt),
		UpdatedAt:         repository.FormatTime(item.UpdatedAt),
	}
}
