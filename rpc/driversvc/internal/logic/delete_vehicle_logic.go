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

// DeleteVehicleLogic 处理删除车辆请求的逻辑结构体。
type DeleteVehicleLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewDeleteVehicleLogic 构造 DeleteVehicleLogic 实例。
func NewDeleteVehicleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteVehicleLogic {
	return &DeleteVehicleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteVehicle 根据车辆 ID 软删除车辆（设置 deleted_at，不物理删除）。
func (l *DeleteVehicleLogic) DeleteVehicle(in *proto.DeleteVehicleRequest) (*proto.DeleteVehicleResponse, error) {
	// 先查询车辆是否存在
	var v model.DriverVehicle
	err := l.svcCtx.DB.First(&v, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("vehicle not found") // 车辆不存在
	}
	if err != nil {
		return nil, err
	}
	// 软删除：GORM 自动设置 deleted_at 字段
	if err := l.svcCtx.DB.Delete(&v).Error; err != nil {
		return nil, err
	}
	// 返回被删除的 ID 与成功标志
	return &proto.DeleteVehicleResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
