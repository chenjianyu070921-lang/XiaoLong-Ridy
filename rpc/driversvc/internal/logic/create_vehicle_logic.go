package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateVehicleLogic 处理创建车辆请求的逻辑结构体。
type CreateVehicleLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewCreateVehicleLogic 构造 CreateVehicleLogic 实例。
func NewCreateVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVehicleLogic {
	return &CreateVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateVehicle 创建车辆记录，状态初始为待审核（1）。
func (l *CreateVehicleLogic) CreateVehicle(in *proto.CreateVehicleRequest) (*proto.CreateVehicleResponse, error) {
	// 注册日期与保险到期日可选，未传则留空
	var regDate *time.Time
	if in.RegistrationDate != "" {
		t, err := time.Parse("2006-01-02", in.RegistrationDate)
		if err != nil {
			return nil, err
		}
		regDate = &t
	}
	var expireAt *time.Time
	if in.InsuranceExpireAt != "" {
		t, err := time.Parse("2006-01-02", in.InsuranceExpireAt)
		if err != nil {
			return nil, err
		}
		expireAt = &t
	}

	v := &model.DriverVehicle{
		DriverId:         uint64(in.DriverId),  // 所属司机 ID
		PlateNo:          in.PlateNo,           // 车牌号
		Brand:            in.Brand,             // 品牌
		Model:            in.Model,             // 车型
		Color:            in.Color,             // 车身颜色
		VehicleType:      int8(in.VehicleType), // 车辆类型
		RegistrationDate: regDate,              // 注册日期
		InsuranceNo:      in.InsuranceNo,       // 保险单号
		InsuranceExpireAt: expireAt,            // 保险到期日
		Status:          1,                     // 初始状态：待审核
	}
	// 写入数据库
	if err := l.svcCtx.DB.Create(v).Error; err != nil {
		return nil, err
	}
	// 返回新建车辆 ID 与初始状态
	return &proto.CreateVehicleResponse{
		Id:     int64(v.Id),
		Status: int32(v.Status),
	}, nil
}
