package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverCertificationLogic {
	return &GetDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询司机审核详情。
func (l *GetDriverCertificationLogic) GetDriverCertification(in *adminsvc.DriverCertificationDetailRequest) (*adminsvc.DriverCertification, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.DriverCertification{}, nil
}
