package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// MenusLogic 处理后台菜单查询 RPC。
type MenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewMenusLogic 创建菜单查询逻辑对象。
func NewMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MenusLogic {
	return &MenusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Menus 根据会话中的角色读取服务配置中的菜单定义。
// token 存在时只信任会话角色，防止客户端伪造 role 参数获取其他角色的菜单。
func (l *MenusLogic) Menus(in *adminsvc.MenusRequest) (*adminsvc.MenusResponse, error) {
	if in.GetToken() != "" {
		admin, err := validateSession(l.ctx, l.svcCtx, in.GetToken())
		if err != nil {
			return nil, err
		}
		return &adminsvc.MenusResponse{Items: mapMenus(admin.Role, l.svcCtx.Config.MenuRoles)}, nil
	}
	return &adminsvc.MenusResponse{Items: mapMenus(in.GetRole(), l.svcCtx.Config.MenuRoles)}, nil
}
