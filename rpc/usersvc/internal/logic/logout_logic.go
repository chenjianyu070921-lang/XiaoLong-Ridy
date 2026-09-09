package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// LogoutLogic 处理用户退出登录 RPC 的业务流程。
type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLogoutLogic 创建退出登录逻辑实例，并绑定请求上下文与服务依赖。
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Logout 废弃当前 Access Token。
func (l *LogoutLogic) Logout(in *userproto.LogoutRequest) (*userproto.LogoutResponse, error) {
	// 当前实现将 Access Token 写入本地注销集合；后续可替换为 Redis 黑名单。
	if err := l.svcCtx.Tokens.Revoke(strings.TrimSpace(in.GetToken())); err != nil {
		return nil, err
	}

	return &userproto.LogoutResponse{Success: true}, nil
}
