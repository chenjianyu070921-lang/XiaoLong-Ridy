package adminservicelogic

import (
	"context"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoginLogic 处理管理员登录 RPC。
type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLoginLogic 创建登录逻辑对象。
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 校验用户名密码，生成 Redis 会话 token，并记录登录日志。
func (l *LoginLogic) Login(in *adminsvc.LoginRequest) (*adminsvc.AuthResponse, error) {
	if strings.TrimSpace(in.GetUsername()) == "" || strings.TrimSpace(in.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "用户名和密码不能为空")
	}
	admin, err := getAdminByUsername(l.ctx, l.svcCtx, in.GetUsername())
	if err != nil {
		return nil, err
	}
	if admin.Status != 1 {
		return nil, status.Error(codes.PermissionDenied, "管理员账号已停用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(in.GetPassword())); err != nil {
		return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	if err := saveSession(l.ctx, l.svcCtx, adminSession{
		AdminID:  admin.ID,
		Username: admin.Username,
		RealName: admin.RealName,
		Role:     admin.Role,
		Status:   admin.Status,
		Token:    token,
	}); err != nil {
		return nil, err
	}
	now := time.Now()
	if _, err := l.svcCtx.MySQL.ExecContext(l.ctx, `UPDATE admin_user SET last_login_at = ?, updated_at = ? WHERE id = ?`, now, now, admin.ID); err != nil {
		return nil, err
	}
	_ = createOperationLog(l.ctx, l.svcCtx, admin.ID, "auth", "login", "admin_user", admin.ID, "管理员登录", "")
	return &adminsvc.AuthResponse{
		Token:     token,
		ExpiresIn: int64(l.svcCtx.Config.Auth.SessionTTLHours) * 3600,
		Admin:     toAdminPB(admin),
	}, nil
}
