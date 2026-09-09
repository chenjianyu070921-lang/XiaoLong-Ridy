package logic

import (
	"context"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// RefreshTokenLogic 处理刷新令牌 RPC 的业务流程。
type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewRefreshTokenLogic 创建刷新令牌逻辑实例，并绑定请求上下文与服务依赖。
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshToken 根据 Refresh Token 轮换新的登录令牌。
func (l *RefreshTokenLogic) RefreshToken(in *userproto.RefreshTokenRequest) (*userproto.RefreshTokenResponse, error) {
	// TokenManager 内部负责旧 Refresh Token 失效和新令牌对签发。
	token, refreshToken, err := l.svcCtx.Tokens.Refresh(strings.TrimSpace(in.GetRefreshToken()))
	if err != nil {
		return nil, err
	}

	return &userproto.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}
