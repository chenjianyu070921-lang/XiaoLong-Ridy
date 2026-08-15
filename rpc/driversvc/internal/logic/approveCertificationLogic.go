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
	return &ApproveCertificationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ApproveCertification 审核通过司机资质。
func (l *ApproveCertificationLogic) ApproveCertification(in *proto.AuditCertificationRequest) (*proto.CommonResponse, error) {
	if err := updateCertificationAudit(l.ctx, l.svcCtx, in.GetCertificationId(), in.GetOperatorId(), in.GetRemark(), 2); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
