package logic

import (
	"context"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVehicleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVehicleLogic {
	return &CreateVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateVehicle 创建车辆，状态初始为待审核（PENDING）。
func (l *CreateVehicleLogic) CreateVehicle(in *proto.CreateVehicleRequest) (*proto.CreateVehicleResponse, error) {
	v := &model.DriverVehicle{
		DriverId:         uint64(in.DriverId),
		PlateNo:          in.PlateNo,
		Brand:            in.Brand,
		Model:            in.Model,
		Color:            in.Color,
		VehicleType:      int8(in.VehicleType),
		InsuranceNo:      in.InsuranceNo,
		Status:           int8(proto.VehicleStatus_VEHICLE_STATUS_PENDING),
	}
	if in.RegistrationDate != nil {
		t := time.Unix(in.GetRegistrationDate(), 0)
		v.RegistrationDate = &t
	}
	if in.InsuranceExpireAt != nil {
		t := time.Unix(in.GetInsuranceExpireAt(), 0)
		v.InsuranceExpireAt = &t
	}
	if err := l.svcCtx.DriverVehicleRepository.Create(l.ctx, v); err != nil {
		return nil, err
	}
	return &proto.CreateVehicleResponse{
		Id:        int64(v.Id),
		Status:    proto.VehicleStatus_VEHICLE_STATUS_PENDING,
		CreatedAt: v.CreatedAt.Unix(),
	}, nil
}
