package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if in == nil || in.GetCertificationId() <= 0 || in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "certification id and operator id are required")
	}
	if in.GetRemark() == "" {
		return nil, status.Error(codes.InvalidArgument, "reject remark is required")
	}
	if l.svcCtx == nil || l.svcCtx.CertificationRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "database not ready")
	}
	if err := l.svcCtx.CertificationRepository.UpdateAudit(l.ctx, in.GetCertificationId(), in.GetOperatorId(), in.GetRemark(), 3); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
