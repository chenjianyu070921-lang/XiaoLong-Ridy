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

// GetVehicleLogic 处理查询车辆详情请求的逻辑结构体。
type GetVehicleLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewGetVehicleLogic 构造 GetVehicleLogic 实例。
func NewGetVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetVehicleLogic {
	return &GetVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetVehicle 根据车辆 ID 查询车辆完整信息。
func (l *GetVehicleLogic) GetVehicle(in *proto.GetVehicleRequest) (*proto.GetVehicleResponse, error) {
	// 按 ID 查询车辆
	var v model.DriverVehicle
	err := l.svcCtx.DB.First(&v, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found") // 车辆不存在
	}
	if err != nil {
		return nil, err
	}
	// 组装并返回车辆详情
	return &proto.GetVehicleResponse{
		Vehicle: &proto.Vehicle{
			Id:                int64(v.Id),                          // 车辆 ID
			DriverId:          int64(v.DriverId),                    // 所属司机 ID
			PlateNo:           v.PlateNo,                           // 车牌号
			Brand:             v.Brand,                             // 品牌
			Model:             v.Model,                             // 车型
			Color:             v.Color,                             // 车身颜色
			VehicleType:       int32(v.VehicleType),                // 车辆类型
			RegistrationDate:  formatDate(v.RegistrationDate),      // 注册日期
			InsuranceNo:       v.InsuranceNo,                       // 保险单号
			InsuranceExpireAt: formatDate(v.InsuranceExpireAt),     // 保险到期日
			Status:            int32(v.Status),                     // 状态
			CreatedAt:         v.CreatedAt.Unix(),                  // 创建时间
			UpdatedAt:         v.UpdatedAt.Unix(),                  // 更新时间
		},
	}, nil
}
