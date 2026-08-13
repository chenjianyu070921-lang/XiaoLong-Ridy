package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateCertificationLogic 处理更新司机认证请求的逻辑结构体。
type UpdateCertificationLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewUpdateCertificationLogic 构造 UpdateCertificationLogic 实例。
func NewUpdateCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCertificationLogic {
	return &UpdateCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateCertification 更新认证信息，仅修改请求中显式传入的字段；审核状态变更时记录审核人/时间。
func (l *UpdateCertificationLogic) UpdateCertification(in *proto.UpdateCertificationRequest) (*proto.UpdateCertificationResponse, error) {
	// 先按 ID 查询认证是否存在
	var c model.DriverCertification
	err := l.svcCtx.DB.First(&c, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("certification not found") // 认证不存在
	}
	if err != nil {
		return nil, err
	}

	// 仅更新显式提供的字段（optional 字段为指针，nil 表示不更新）
	updates := map[string]interface{}{}
	if in.DriverId != nil {
		updates["driver_id"] = in.GetDriverId() // 所属司机 ID
	}
	if in.VehicleId != nil {
		updates["vehicle_id"] = in.GetVehicleId() // 关联车辆 ID
	}
	if in.IdCardFrontUrl != nil {
		updates["id_card_front_url"] = in.GetIdCardFrontUrl() // 身份证人像面
	}
	if in.IdCardBackUrl != nil {
		updates["id_card_back_url"] = in.GetIdCardBackUrl() // 身份证国徽面
	}
	if in.DriverLicenseUrl != nil {
		updates["driver_license_url"] = in.GetDriverLicenseUrl() // 驾驶证照片
	}
	if in.VehicleLicenseUrl != nil {
		updates["vehicle_license_url"] = in.GetVehicleLicenseUrl() // 行驶证照片
	}
	if in.AuditStatus != nil {
		updates["audit_status"] = int8(in.GetAuditStatus()) // 审核状态
	}
	if in.AuditRemark != nil {
		updates["audit_remark"] = in.GetAuditRemark() // 驳回原因/审核备注
	}
	if in.AuditedBy != nil {
		updates["audited_by"] = in.GetAuditedBy() // 审核人
	}
	if in.AuditedAt != nil {
		t := time.Unix(in.GetAuditedAt(), 0)
		updates["audited_at"] = &t // 审核时间
	}

	// 执行更新
	if err := l.svcCtx.DB.Model(&c).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新读取更新后的记录
	if err := l.svcCtx.DB.First(&c, in.Id).Error; err != nil {
		return nil, err
	}
	// 返回更新后的 ID、审核状态与更新时间
	return &proto.UpdateCertificationResponse{
		Id:          int64(c.Id),
		AuditStatus: int32(c.AuditStatus),
		UpdatedAt:   c.UpdatedAt.Unix(),
	}, nil
}
