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

// DeleteScoreLogic 处理删除服务分请求的逻辑结构体。
type DeleteScoreLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewDeleteScoreLogic 构造 DeleteScoreLogic 实例。
func NewDeleteScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteScoreLogic {
	return &DeleteScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteScore 根据记录 ID 软删除服务分（设置 deleted_at，不物理删除）。
func (l *DeleteScoreLogic) DeleteScore(in *proto.DeleteScoreRequest) (*proto.DeleteScoreResponse, error) {
	// 先查询记录是否存在
	var s model.DriverScore
	err := l.svcCtx.DB.First(&s, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("score not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}
	// 软删除：GORM 自动设置 deleted_at 字段
	if err := l.svcCtx.DB.Delete(&s).Error; err != nil {
		return nil, err
	}
	// 返回被删除的 ID 与成功标志
	return &proto.DeleteScoreResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
