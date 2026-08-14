package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// RejectDriverCertificationLogic 处理司机认证驳回 RPC。
type RejectDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewRejectDriverCertificationLogic 创建审核驳回逻辑对象。
func NewRejectDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectDriverCertificationLogic {
	return &RejectDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RejectDriverCertification 驳回司机认证，只更新审核记录并写入操作日志。
func (l *RejectDriverCertificationLogic) RejectDriverCertification(in *adminsvc.AuditDriverCertificationRequest) (*adminsvc.CommonResponse, error) {
	if err := auditCertification(l.ctx, l.svcCtx, in, 3); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
