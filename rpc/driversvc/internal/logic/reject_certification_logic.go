package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectCertificationLogic {
	return &RejectCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核驳回司机资质。
func (l *RejectCertificationLogic) RejectCertification(in *__proto.AuditCertificationRequest) (*__proto.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &__proto.CommonResponse{}, nil
}
