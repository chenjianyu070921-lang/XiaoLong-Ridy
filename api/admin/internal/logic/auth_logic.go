package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/model"
	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
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

// AuthLogic 封装 HTTP 层到 adminsvc 鉴权 RPC 的适配逻辑。
// api/admin 只负责参数校验、token 透传和响应结构映射，不再直接读写 admin_user 或 Redis 会话。
type AuthLogic struct {
	ctx *svc.ServiceContext
}

// NewAuthLogic 创建鉴权适配逻辑对象。
func NewAuthLogic(ctx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{ctx: ctx}
}

// Register 注册后台管理员。
// 首个管理员是否允许免登录、后续是否要求超级管理员，均由 adminsvc 统一判断。
func (l *AuthLogic) Register(ctx context.Context, req *types.RegisterRequest, current *model.AdminSession) (*types.AuthResponse, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.RealName) == "" {
		return nil, ErrBadRequest
	}
	if req.Role <= 0 {
		return nil, ErrBadRequest
	}
	rpcReq := &adminclient.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Role:     req.Role,
	}
	if current != nil {
		rpcReq.OperatorAdminId = current.AdminID
		rpcReq.OperatorRole = current.Role
	}
	resp, err := l.ctx.AdminSvc.Register(ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return toAuthResponse(resp), nil
}

// Login 透传管理员账号密码到 adminsvc，由 adminsvc 完成密码校验和 Redis 会话创建。
func (l *AuthLogic) Login(ctx context.Context, req *types.LoginRequest) (*types.AuthResponse, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrBadRequest
	}
	resp, err := l.ctx.AdminSvc.Login(ctx, &adminclient.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return toAuthResponse(resp), nil
}

// Logout 将 token 透传给 adminsvc，由 adminsvc 删除 Redis 会话。
func (l *AuthLogic) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrUnauthorized
	}
	_, err := l.ctx.AdminSvc.Logout(ctx, &adminclient.LogoutRequest{Token: token})
	return err
}

// ValidateSession 校验 token 并返回可放入 HTTP request context 的管理员会话快照。
func (l *AuthLogic) ValidateSession(ctx context.Context, token string) (*model.AdminSession, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrUnauthorized
	}
	resp, err := l.ctx.AdminSvc.ValidateSession(ctx, &adminclient.ValidateSessionRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return toAdminSession(token, resp.GetAdmin()), nil
}

// Me 透传 token 到 adminsvc 查询当前管理员信息。
func (l *AuthLogic) Me(ctx context.Context, token string) (*types.MeResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrUnauthorized
	}
	resp, err := l.ctx.AdminSvc.Me(ctx, &adminclient.MeRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return &types.MeResponse{Admin: toAdminDTO(resp.GetAdmin())}, nil
}

// GetMenus 透传 token 到 adminsvc 查询当前管理员角色菜单。
func (l *AuthLogic) GetMenus(ctx context.Context, token string) ([]types.MenuItem, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrUnauthorized
	}
	resp, err := l.ctx.AdminSvc.Menus(ctx, &adminclient.MenusRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return toMenuItems(resp.GetItems()), nil
}

// toAuthResponse 将 adminsvc 登录/注册响应转换为 HTTP 对外响应。
func toAuthResponse(resp *adminclient.AuthResponse) *types.AuthResponse {
	if resp == nil {
		return &types.AuthResponse{}
	}
	return &types.AuthResponse{
		Token:     resp.GetToken(),
		ExpiresIn: resp.GetExpiresIn(),
		Admin:     toAdminDTO(resp.GetAdmin()),
	}
}

// toAdminDTO 将 adminsvc 管理员对象转换为 HTTP DTO。
func toAdminDTO(admin *adminclient.Admin) types.AdminDTO {
	if admin == nil {
		return types.AdminDTO{}
	}
	return types.AdminDTO{
		ID:       admin.GetId(),
		Username: admin.GetUsername(),
		RealName: admin.GetRealName(),
		Role:     admin.GetRole(),
		Status:   admin.GetStatus(),
	}
}

// toAdminSession 将 adminsvc 返回的管理员对象转换为 HTTP context 使用的会话快照。
func toAdminSession(token string, admin *adminclient.Admin) *model.AdminSession {
	if admin == nil {
		return nil
	}
	return &model.AdminSession{
		Token:    token,
		AdminID:  admin.GetId(),
		Username: admin.GetUsername(),
		RealName: admin.GetRealName(),
		Role:     admin.GetRole(),
	}
}

// toMenuItems 将 adminsvc 菜单树递归转换为 HTTP 返回结构。
func toMenuItems(items []*adminclient.MenuItem) []types.MenuItem {
	result := make([]types.MenuItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, types.MenuItem{
			Name:     item.GetName(),
			Path:     item.GetPath(),
			Icon:     item.GetIcon(),
			Perm:     item.GetPerm(),
			Children: toMenuItems(item.GetChildren()),
		})
	}
	return result
}
