package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDriverLogic {
	return &DeleteDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteDriver 根据司机 ID 软删除司机（设置 deleted_at，不物理删除）。
func (l *DeleteDriverLogic) DeleteDriver(in *proto.DeleteDriverRequest) (*proto.DeleteDriverResponse, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	d, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.DriverRepository.Delete(l.ctx, d); err != nil {
		return nil, err
	}
	return &proto.DeleteDriverResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
