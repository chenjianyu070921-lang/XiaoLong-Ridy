package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDriverCertificationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDriverCertificationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriverCertificationsLogic {
	return &ListDriverCertificationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询司机审核列表。
func (l *ListDriverCertificationsLogic) ListDriverCertifications(in *adminsvc.DriverCertificationListRequest) (*adminsvc.DriverCertificationListResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.DriverCertificationListResponse{}, nil
}
