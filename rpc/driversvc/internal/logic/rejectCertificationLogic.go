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
	return &RejectCertificationLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// RejectCertification 驳回司机资质审核。
func (l *RejectCertificationLogic) RejectCertification(in *proto.AuditCertificationRequest) (*proto.CommonResponse, error) {
	if err := updateCertificationAudit(l.ctx, l.svcCtx, in.GetCertificationId(), in.GetOperatorId(), in.GetRemark(), 3); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
