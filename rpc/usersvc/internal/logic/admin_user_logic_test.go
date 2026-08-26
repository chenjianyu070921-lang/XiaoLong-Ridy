package logic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/usersvc/internal/config"
	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestAdminGetUser_ReturnsUser 验证管理后台详情 RPC 能按真实用户 ID 查询 usersvc 数据。
func TestAdminGetUser_ReturnsUser(t *testing.T) {
	users := repository.NewMemoryUserRepository()
	user := &model.User{Phone: "13800138000", Nickname: "乘客", Status: 1, RegisterSource: "h5"}
	if err := users.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	svcCtx := svc.NewServiceContext(config.Config{}, users, repository.NewMemoryAddressRepository(), repository.NewMemoryCouponRepository(), repository.NewMemoryRiskBlacklistRepository(), nil, nil, nil, nil)
	resp, err := NewAdminUserLogic(context.Background(), svcCtx).GetUser(&userproto.AdminUserDetailRequest{Id: user.ID})
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if resp.GetId() != user.ID || resp.GetPhone() != user.Phone || resp.GetRegisterSource() != "h5" {
		t.Fatalf("GetUser() response = %+v", resp)
	}
}

// TestAdminGetUser_NotFound 验证用户不存在时返回标准 NotFound，便于管理后台映射 404。
func TestAdminGetUser_NotFound(t *testing.T) {
	users := repository.NewMemoryUserRepository()
	svcCtx := svc.NewServiceContext(config.Config{}, users, repository.NewMemoryAddressRepository(), repository.NewMemoryCouponRepository(), repository.NewMemoryRiskBlacklistRepository(), nil, nil, nil, nil)
	_, err := NewAdminUserLogic(context.Background(), svcCtx).GetUser(&userproto.AdminUserDetailRequest{Id: 404})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetUser() code = %v, want NotFound", status.Code(err))
	}
}
