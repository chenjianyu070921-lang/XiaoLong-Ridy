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

// DeleteWithdrawLogic 处理删除提现请求的逻辑结构体。
type DeleteWithdrawLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewDeleteWithdrawLogic 构造 DeleteWithdrawLogic 实例。
func NewDeleteWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteWithdrawLogic {
	return &DeleteWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteWithdraw 根据提现 ID 软删除提现记录（设置 deleted_at，不物理删除）。
func (l *DeleteWithdrawLogic) DeleteWithdraw(in *proto.DeleteWithdrawRequest) (*proto.DeleteWithdrawResponse, error) {
	// 先查询记录是否存在
	var w model.DriverWithdraw
	err := l.svcCtx.DB.First(&w, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("withdraw not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}
	// 软删除：GORM 自动设置 deleted_at 字段
	if err := l.svcCtx.DB.Delete(&w).Error; err != nil {
		return nil, err
	}
	// 返回被删除的 ID 与成功标志
	return &proto.DeleteWithdrawResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
