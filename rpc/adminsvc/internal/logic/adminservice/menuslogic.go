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

// Menus 返回 P0 阶段按角色固定配置的后台菜单。
func (l *MenusLogic) Menus(in *adminsvc.MenusRequest) (*adminsvc.MenusResponse, error) {
	if in.GetToken() != "" {
		admin, err := validateSession(l.ctx, l.svcCtx, in.GetToken())
		if err != nil {
			return nil, err
		}
		return &adminsvc.MenusResponse{Items: mapMenus(admin.Role)}, nil
	}
	return &adminsvc.MenusResponse{Items: mapMenus(in.GetRole())}, nil
}
