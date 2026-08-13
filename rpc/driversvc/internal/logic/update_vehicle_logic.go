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

// UpdateVehicleLogic 处理更新车辆请求的逻辑结构体。
type UpdateVehicleLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewUpdateVehicleLogic 构造 UpdateVehicleLogic 实例。
func NewUpdateVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateVehicleLogic {
	return &UpdateVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateVehicle 更新车辆信息，仅修改请求中显式传入的字段。
func (l *UpdateVehicleLogic) UpdateVehicle(in *proto.UpdateVehicleRequest) (*proto.UpdateVehicleResponse, error) {
	// 先按 ID 查询车辆是否存在
	var v model.DriverVehicle
	err := l.svcCtx.DB.First(&v, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found") // 车辆不存在
	}
	if err != nil {
		return nil, err
	}

	// 仅更新显式提供的字段（optional 字段为指针，nil 表示不更新）
	updates := map[string]interface{}{}
	if in.DriverId != nil {
		updates["driver_id"] = in.GetDriverId() // 所属司机 ID
	}
	if in.PlateNo != nil {
		updates["plate_no"] = in.GetPlateNo() // 车牌号
	}
	if in.Brand != nil {
		updates["brand"] = in.GetBrand() // 品牌
	}
	if in.Model != nil {
		updates["model"] = in.GetModel() // 车型
	}
	if in.Color != nil {
		updates["color"] = in.GetColor() // 车身颜色
	}
	if in.VehicleType != nil {
		updates["vehicle_type"] = int8(in.GetVehicleType()) // 车辆类型
	}
	if in.RegistrationDate != nil {
		if t, perr := time.Parse("2006-01-02", in.GetRegistrationDate()); perr == nil {
			updates["registration_date"] = t // 注册日期
		}
	}
	if in.InsuranceNo != nil {
		updates["insurance_no"] = in.GetInsuranceNo() // 保险单号
	}
	if in.InsuranceExpireAt != nil {
		if t, perr := time.Parse("2006-01-02", in.GetInsuranceExpireAt()); perr == nil {
			updates["insurance_expire_at"] = t // 保险到期日
		}
	}
	if in.Status != nil {
		updates["status"] = int8(in.GetStatus()) // 状态
	}

	// 执行更新
	if err := l.svcCtx.DB.Model(&v).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新读取更新后的记录
	if err := l.svcCtx.DB.First(&v, in.Id).Error; err != nil {
		return nil, err
	}
	// 返回更新后的 ID、状态与更新时间
	return &proto.UpdateVehicleResponse{
		Id:        int64(v.Id),
		Status:    int32(v.Status),
		UpdatedAt: v.UpdatedAt.Unix(),
	}, nil
}
