package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// GetCertificationLogic 处理查询认证详情请求的逻辑结构体。
type GetCertificationLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewGetCertificationLogic 构造 GetCertificationLogic 实例。
func NewGetCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCertificationLogic {
	return &GetCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetCertification 根据认证 ID 查询认证完整信息。
func (l *GetCertificationLogic) GetCertification(in *proto.GetCertificationRequest) (*proto.GetCertificationResponse, error) {
	// 按 ID 查询认证
	var c model.DriverCertification
	err := l.svcCtx.DB.First(&c, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("certification not found") // 认证不存在
	}
	if err != nil {
		return nil, err
	}
	// 组装并返回认证详情
	return &proto.GetCertificationResponse{
		Certification: &proto.Certification{
			Id:                int64(c.Id),                     // 认证 ID
			DriverId:          int64(c.DriverId),               // 所属司机 ID
			VehicleId:         int64(c.VehicleId),              // 关联车辆 ID
			IdCardFrontUrl:    c.IdCardFrontUrl,                // 身份证人像面
			IdCardBackUrl:     c.IdCardBackUrl,                 // 身份证国徽面
			DriverLicenseUrl:  c.DriverLicenseUrl,              // 驾驶证照片
			VehicleLicenseUrl: c.VehicleLicenseUrl,             // 行驶证照片
			AuditStatus:       int32(c.AuditStatus),            // 审核状态
			AuditRemark:       c.AuditRemark,                   // 驳回原因/审核备注
			AuditedBy:         int64(c.AuditedBy),              // 审核人
			AuditedAt:         unixOrZero(c.AuditedAt),         // 审核时间
			CreatedAt:         c.CreatedAt.Unix(),              // 创建时间
			UpdatedAt:         c.UpdatedAt.Unix(),              // 更新时间
		},
	}, nil
}
