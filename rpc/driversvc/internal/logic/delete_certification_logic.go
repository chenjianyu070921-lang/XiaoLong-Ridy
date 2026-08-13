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

// DeleteCertificationLogic 处理删除认证请求的逻辑结构体。
type DeleteCertificationLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewDeleteCertificationLogic 构造 DeleteCertificationLogic 实例。
func NewDeleteCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCertificationLogic {
	return &DeleteCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteCertification 根据认证 ID 软删除认证记录（设置 deleted_at，不物理删除）。
func (l *DeleteCertificationLogic) DeleteCertification(in *proto.DeleteCertificationRequest) (*proto.DeleteCertificationResponse, error) {
	// 先查询认证是否存在
	var c model.DriverCertification
	err := l.svcCtx.DB.First(&c, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("certification not found") // 认证不存在
	}
	if err != nil {
		return nil, err
	}
	// 软删除：GORM 自动设置 deleted_at 字段
	if err := l.svcCtx.DB.Delete(&c).Error; err != nil {
		return nil, err
	}
	// 返回被删除的 ID 与成功标志
	return &proto.DeleteCertificationResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
