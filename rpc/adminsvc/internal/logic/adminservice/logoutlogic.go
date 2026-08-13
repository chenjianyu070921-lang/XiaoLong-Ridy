package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// LogoutLogic 处理管理员退出登录 RPC。
type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLogoutLogic 创建退出登录逻辑对象。
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Logout 删除 Redis 会话，使当前 token 失效。
func (l *LogoutLogic) Logout(in *adminsvc.LogoutRequest) (*adminsvc.CommonResponse, error) {
	if err := deleteSession(l.ctx, l.svcCtx, in.GetToken()); err != nil {
		return nil, err
	}
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}
