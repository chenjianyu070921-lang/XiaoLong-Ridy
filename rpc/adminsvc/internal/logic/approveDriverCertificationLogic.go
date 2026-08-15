package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveDriverCertificationLogic {
	return &ApproveDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核通过司机认证。
func (l *ApproveDriverCertificationLogic) ApproveDriverCertification(in *adminsvc.AuditDriverCertificationRequest) (*adminsvc.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.CommonResponse{}, nil
}
