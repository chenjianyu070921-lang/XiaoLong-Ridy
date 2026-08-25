package adminservicelogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 登录防爆破参数：同一用户名连续失败达到阈值后锁定一段时间，
// 配合 Redis 计数实现，避免在线暴力破解后台账号。
const (
	loginFailKeyPrefix = "admin:login:fail:"
	maxLoginFailures   = 5
	loginLockDuration  = 15 * time.Minute
)

// loginFailKey 返回记录指定用户名登录失败次数的 Redis key。
func loginFailKey(username string) string {
	return loginFailKeyPrefix + username
}

// loginFailureExceeded 判断累计失败次数是否已达到锁定阈值。
func loginFailureExceeded(failures int64) bool {
	return failures >= maxLoginFailures
}

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
	username := strings.TrimSpace(in.GetUsername())
	// 已锁定的账号直接拒绝，避免继续执行耗时的 bcrypt 哈希比对，降低爆破速率。
	if err := l.checkLoginLock(username); err != nil {
		return nil, err
	}
	admin, err := getAdminByUsername(l.ctx, l.svcCtx, username)
	if err != nil {
		// 账号不存在与密码错误返回同一提示，避免枚举有效用户名；
		// 不存在的用户名同样计入失败次数，防止针对任意用户名的爆破。
		if status.Code(err) == codes.NotFound {
			l.recordLoginFailure(username)
			return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
		}
		return nil, err
	}
	if admin.Status != 1 {
		return nil, status.Error(codes.PermissionDenied, "管理员账号已停用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(in.GetPassword())); err != nil {
		l.recordLoginFailure(username)
		return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
	}
	// 登录成功后清除失败计数，避免历史失败影响后续正常登录。
	l.clearLoginFailure(username)
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
		ExpiresIn: int64(l.svcCtx.Config.Session.SessionTTLHours) * 3600,
		Admin:     toAdminPB(admin),
	}, nil
}

// checkLoginLock 检查指定用户名是否已达到登录失败锁定阈值。
// Redis 读取异常时返回错误，交由上层统一处理；未达到阈值时返回 nil。
func (l *LoginLogic) checkLoginLock(username string) error {
	count, err := l.svcCtx.Redis.Get(l.ctx, loginFailKey(username)).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if loginFailureExceeded(count) {
		return status.Error(codes.PermissionDenied, "登录失败次数过多，请稍后重试")
	}
	return nil
}

// recordLoginFailure 累加指定用户名的登录失败次数，并在首次失败时设置过期时间，
// 使锁定窗口到期后自动回收，无需人工解锁。
func (l *LoginLogic) recordLoginFailure(username string) {
	key := loginFailKey(username)
	count, err := l.svcCtx.Redis.Incr(l.ctx, key).Result()
	if err != nil {
		l.Logger.Errorf("record login failure: %v", err)
		return
	}
	if count == 1 {
		_ = l.svcCtx.Redis.Expire(l.ctx, key, loginLockDuration).Err()
	}
}

// clearLoginFailure 删除指定用户名的登录失败计数，登录成功后调用。
func (l *LoginLogic) clearLoginFailure(username string) {
	_ = l.svcCtx.Redis.Del(l.ctx, loginFailKey(username)).Err()
}
