package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// ValidateSessionLogic 处理管理员会话校验 RPC。
// api/admin 的鉴权中间件只透传 token，Redis 会话读取和账号状态校验统一由 adminsvc 负责。
type ValidateSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewValidateSessionLogic 创建管理员会话校验逻辑对象。
func NewValidateSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateSessionLogic {
	return &ValidateSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ValidateSession 校验 token 并返回当前有效管理员信息。
func (l *ValidateSessionLogic) ValidateSession(in *adminsvc.ValidateSessionRequest) (*adminsvc.ValidateSessionResponse, error) {
	admin, err := validateSession(l.ctx, l.svcCtx, in.GetToken())
	if err != nil {
		return nil, err
	}
	return &adminsvc.ValidateSessionResponse{Admin: toAdminPB(admin)}, nil
}
