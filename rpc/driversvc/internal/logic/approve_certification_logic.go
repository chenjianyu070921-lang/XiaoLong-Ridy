package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveCertificationLogic {
	return &ApproveCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核通过司机资质。
func (l *ApproveCertificationLogic) ApproveCertification(in *__proto.AuditCertificationRequest) (*__proto.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &__proto.CommonResponse{}, nil
}
