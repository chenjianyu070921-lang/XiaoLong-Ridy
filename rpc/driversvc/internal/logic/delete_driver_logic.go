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
	var d model.Driver
	err := l.svcCtx.DB.First(&d, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("driver not found")
	}
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.DB.Delete(&d).Error; err != nil {
		return nil, err
	}
	return &proto.DeleteDriverResponse{
		Id:      in.Id,
		Success: true,
	}, nil
}
