package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if in == nil || in.GetCertificationId() <= 0 || in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "certification id and operator id are required")
	}
	if l.svcCtx == nil || l.svcCtx.CertificationRepository == nil {
		return nil, status.Error(codes.FailedPrecondition, "database not ready")
	}
	if err := l.svcCtx.CertificationRepository.UpdateAudit(l.ctx, in.GetCertificationId(), in.GetOperatorId(), in.GetRemark(), 2); err != nil {
		return nil, err
	}
	return &proto.CommonResponse{Message: "ok"}, nil
}
