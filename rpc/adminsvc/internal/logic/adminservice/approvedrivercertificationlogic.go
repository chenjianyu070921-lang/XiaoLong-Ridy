package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ApproveDriverCertificationLogic 处理司机认证审核通过 RPC。
type ApproveDriverCertificationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewApproveDriverCertificationLogic 创建审核通过逻辑对象。
func NewApproveDriverCertificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveDriverCertificationLogic {
	return &ApproveDriverCertificationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ApproveDriverCertification 审核通过司机认证，并同步激活司机和车辆。
func (l *ApproveDriverCertificationLogic) ApproveDriverCertification(in *adminsvc.AuditDriverCertificationRequest) (*adminsvc.CommonResponse, error) {
	if err := auditCertification(l.ctx, l.svcCtx, in, 2); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
