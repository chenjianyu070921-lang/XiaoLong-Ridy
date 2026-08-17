package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectDriverCertificationLogic {
	return &RejectDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 驳回司机认证。
func (l *RejectDriverCertificationLogic) RejectDriverCertification(in *adminsvc.AuditDriverCertificationRequest) (*adminsvc.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.CommonResponse{}, nil
}
