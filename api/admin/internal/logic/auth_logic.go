package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/repository"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrUnauthorized 表示未登录、token 无效或账号密码错误。
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden 表示已登录但权限不足或账号被禁用。
	ErrForbidden = errors.New("forbidden")
	// ErrConflict 表示资源冲突，例如用户名已存在。
	ErrConflict = errors.New("conflict")
	// ErrBadRequest 表示请求参数不合法。
	ErrBadRequest = errors.New("bad request")
)

// AuthLogic 封装管理员注册、登录、退出和菜单权限逻辑。
// handler 层只负责 HTTP 入参出参，真正的鉴权业务规则集中在这里维护。
type AuthLogic struct {
	ctx *svc.ServiceContext
}

// NewAuthLogic 创建鉴权业务逻辑对象。
func NewAuthLogic(ctx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{ctx: ctx}
}

// Register 注册后台管理员。
// 系统内还没有管理员时允许免登录创建首个超级管理员；已有管理员后，只允许超级管理员继续创建账号。
func (l *AuthLogic) Register(ctx context.Context, req *types.RegisterRequest, current *model.AdminSession) (*types.AuthResponse, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrBadRequest
	}
	if strings.TrimSpace(req.RealName) == "" {
		return nil, ErrBadRequest
	}
	if req.Role < model.AdminRoleSuper || req.Role > model.AdminRoleCS {
		return nil, ErrBadRequest
	}

	// 首个管理员用于系统初始化；后续注册必须走超级管理员授权。
	count, err := l.ctx.AdminRepository.Count(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		if current == nil {
			return nil, ErrUnauthorized
		}
		if current.Role != model.AdminRoleSuper {
			return nil, ErrForbidden
		}
	}

	// 用户名唯一性由数据库唯一索引兜底，这里提前查询是为了返回更清晰的业务错误。
	existing, err := l.ctx.AdminRepository.GetByUsername(ctx, req.Username)
	if err == nil && existing != nil {
		return nil, ErrConflict
	}
	if err != nil && !errors.Is(err, repository.ErrAdminNotFound) {
		return nil, err
	}

	// 密码必须使用 bcrypt 哈希后入库，不能保存明文密码。
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	adminID, err := l.ctx.AdminRepository.Create(ctx, repository.CreateAdminInput{
		Username:     req.Username,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Role:         req.Role,
		Status:       model.AdminStatusNormal,
	})
	if err != nil {
		return nil, err
	}

	admin, err := l.ctx.AdminRepository.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return l.buildAuthResponse(ctx, admin)
}

// Login 校验管理员账号密码，并创建 Redis 会话。
func (l *AuthLogic) Login(ctx context.Context, req *types.LoginRequest) (*types.AuthResponse, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrBadRequest
	}
	admin, err := l.ctx.AdminRepository.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrAdminNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if admin.Status != model.AdminStatusNormal {
		return nil, ErrForbidden
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrUnauthorized
	}

	now := time.Now()
	if err := l.ctx.AdminRepository.UpdateLastLoginAt(ctx, admin.ID, now); err != nil {
		return nil, err
	}
	admin.LastLoginAt = &now
	return l.buildAuthResponse(ctx, admin)
}

// Logout 删除 Redis 中的会话，使 token 立即失效。
func (l *AuthLogic) Logout(ctx context.Context, token string) error {
	if token == "" {
		return ErrUnauthorized
	}
	return l.ctx.SessionRepository.Delete(ctx, token)
}

// Me 查询当前登录管理员信息。
func (l *AuthLogic) Me(ctx context.Context, session *model.AdminSession) (*types.MeResponse, error) {
	admin, err := l.ctx.AdminRepository.GetByID(ctx, session.AdminID)
	if err != nil {
		return nil, err
	}
	return &types.MeResponse{Admin: toAdminDTO(admin)}, nil
}

// GetMenus 根据管理员角色返回菜单权限。
// P0 阶段采用代码固定菜单，后续可扩展为权限表配置。
func (l *AuthLogic) GetMenus(session *model.AdminSession) []types.MenuItem {
	common := []types.MenuItem{
		{Name: "工作台", Path: "/dashboard", Perm: "dashboard:view"},
		{Name: "操作日志", Path: "/operation-logs", Perm: "log:view"},
	}
	switch session.Role {
	case model.AdminRoleSuper:
		return []types.MenuItem{
			{Name: "工作台", Path: "/dashboard", Perm: "dashboard:view"},
			{Name: "用户管理", Path: "/users", Perm: "user:view"},
			{Name: "司机管理", Path: "/drivers", Perm: "driver:view"},
			{Name: "订单管理", Path: "/orders", Perm: "order:view"},
			{Name: "营销管理", Path: "/marketing", Perm: "coupon:view"},
			{Name: "风控管理", Path: "/risk", Perm: "risk:view"},
			{Name: "系统管理", Path: "/system", Perm: "system:view"},
		}
	case model.AdminRoleOps:
		return []types.MenuItem{
			{Name: "工作台", Path: "/dashboard", Perm: "dashboard:view"},
			{Name: "用户管理", Path: "/users", Perm: "user:view"},
			{Name: "司机管理", Path: "/drivers", Perm: "driver:view"},
			{Name: "订单管理", Path: "/orders", Perm: "order:view"},
			{Name: "营销管理", Path: "/marketing", Perm: "coupon:view"},
			{Name: "操作日志", Path: "/operation-logs", Perm: "log:view"},
		}
	case model.AdminRoleCS:
		return []types.MenuItem{
			{Name: "工作台", Path: "/dashboard", Perm: "dashboard:view"},
			{Name: "用户管理", Path: "/users", Perm: "user:view"},
			{Name: "订单管理", Path: "/orders", Perm: "order:view"},
			{Name: "操作日志", Path: "/operation-logs", Perm: "log:view"},
		}
	default:
		return common
	}
}

// buildAuthResponse 生成 token、保存 Redis 会话并构造统一鉴权返回。
func (l *AuthLogic) buildAuthResponse(ctx context.Context, admin *model.AdminUser) (*types.AuthResponse, error) {
	token, session, err := l.issueSession(admin)
	if err != nil {
		return nil, err
	}
	if err := l.ctx.SessionRepository.Save(ctx, *session); err != nil {
		return nil, err
	}
	return &types.AuthResponse{
		Token:     token,
		ExpiresIn: int64(l.sessionTTL().Seconds()),
		Admin:     toAdminDTO(admin),
	}, nil
}

// issueSession 生成随机 token 和 Redis 会话内容。
func (l *AuthLogic) issueSession(admin *model.AdminUser) (string, *model.AdminSession, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", nil, err
	}
	now := time.Now()
	session := &model.AdminSession{
		Token:     token,
		AdminID:   admin.ID,
		Username:  admin.Username,
		RealName:  admin.RealName,
		Role:      admin.Role,
		IssuedAt:  now,
		ExpiresAt: now.Add(time.Duration(l.ctx.Config.Auth.SessionTTLHours) * time.Hour),
	}
	return token, session, nil
}

// toAdminDTO 将管理员模型转换为对外返回结构。
func toAdminDTO(admin *model.AdminUser) types.AdminDTO {
	return types.AdminDTO{
		ID:       admin.ID,
		Username: admin.Username,
		RealName: admin.RealName,
		Role:     admin.Role,
		Status:   admin.Status,
	}
}

// randomToken 生成安全随机 token。
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// sessionTTL 返回当前配置中的会话有效期。
func (l *AuthLogic) sessionTTL() time.Duration {
	return time.Duration(l.ctx.Config.Auth.SessionTTLHours) * time.Hour
}
