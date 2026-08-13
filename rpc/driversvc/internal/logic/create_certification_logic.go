package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateCertificationLogic 处理创建司机认证请求的逻辑结构体。
type CreateCertificationLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewCreateCertificationLogic 构造 CreateCertificationLogic 实例。
func NewCreateCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCertificationLogic {
	return &CreateCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateCertification 创建认证记录，审核状态初始为待审核（1）。
func (l *CreateCertificationLogic) CreateCertification(in *proto.CreateCertificationRequest) (*proto.CreateCertificationResponse, error) {
	c := &model.DriverCertification{
		DriverId:          uint64(in.DriverId),    // 所属司机 ID
		VehicleId:         uint64(in.VehicleId),   // 关联车辆 ID
		IdCardFrontUrl:    in.IdCardFrontUrl,      // 身份证人像面
		IdCardBackUrl:     in.IdCardBackUrl,       // 身份证国徽面
		DriverLicenseUrl:  in.DriverLicenseUrl,    // 驾驶证照片
		VehicleLicenseUrl: in.VehicleLicenseUrl,   // 行驶证照片
		AuditStatus:       1,                      // 初始审核状态：待审核
	}
	// 写入数据库
	if err := l.svcCtx.DB.Create(c).Error; err != nil {
		return nil, err
	}
	// 返回新建认证 ID 与初始审核状态
	return &proto.CreateCertificationResponse{
		Id:          int64(c.Id),
		AuditStatus: int32(c.AuditStatus),
	}, nil
}
