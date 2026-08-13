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

// DeleteDriverLogic 处理删除司机请求的逻辑结构体。
type DeleteDriverLogic struct {
	ctx    context.Context      // ctx：请求上下文
	svcCtx *svc.ServiceContext  // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewDeleteDriverLogic 构造 DeleteDriverLogic 实例。
func NewDeleteDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDriverLogic {
	return &DeleteDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteDriver 根据司机 ID 软删除司机（设置 deleted_at，不物理删除）。
func (l *DeleteDriverLogic) DeleteDriver(in *proto.DeleteDriverRequest) (*proto.DeleteDriverResponse, error) {
	// 先查询司机是否存在
	var d model.Driver
	err := l.svcCtx.DB.First(&d, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("driver not found") // 司机不存在
	}
	if err != nil {
		return nil, err
	}
	// 软删除：GORM 自动设置 deleted_at 字段
	if err := l.svcCtx.DB.Delete(&d).Error; err != nil {
		return nil, err
	}
	// 返回被删除的 ID 与成功标志
	return &proto.DeleteDriverResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
