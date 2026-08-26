package adminservicelogic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	usersvc "XiaoLong-Ridy/rpc/usersvc/proto"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// fakeUsersClient 提供用户列表 RPC 的最小测试替身。
type fakeUsersClient struct {
	usersvc.UserClient
	listRequest   *usersvc.AdminUserListRequest
	detailRequest *usersvc.AdminUserDetailRequest
}

// AdminListUsers 记录后台转发的分页和状态条件。
func (f *fakeUsersClient) AdminListUsers(_ context.Context, in *usersvc.AdminUserListRequest, _ ...grpc.CallOption) (*usersvc.AdminUserListResponse, error) {
	f.listRequest = in
	return &usersvc.AdminUserListResponse{List: []*usersvc.AdminUser{{Id: 1001, Phone: "13800138000", Nickname: "乘客", IdCardNo: "110101199001011234"}}, Total: 1, Page: 1, PageSize: 20}, nil
}

// AdminGetUser 记录后台用户详情查询的用户 ID。
func (f *fakeUsersClient) AdminGetUser(_ context.Context, in *usersvc.AdminUserDetailRequest, _ ...grpc.CallOption) (*usersvc.AdminUser, error) {
	f.detailRequest = in
	return &usersvc.AdminUser{Id: in.GetId(), Phone: "13800138000", Nickname: "乘客详情", RealName: "张三", IdCardNo: "110101199001011234", Status: 1}, nil
}

// TestListUsers_UsesUsersRPCWhenFiltersAreSupported 验证用户列表优先查询真实 usersvc。
func TestListUsers_UsesUsersRPCWhenFiltersAreSupported(t *testing.T) {
	client := &fakeUsersClient{}
	logic := NewListUsersLogic(context.Background(), &svc.ServiceContext{UsersSvc: client})
	resp, err := logic.ListUsers(&adminsvc.UserListRequest{Page: 1, PageSize: 20, Status: 1})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if client.listRequest == nil || client.listRequest.GetStatus() != 1 {
		t.Fatalf("usersvc request = %+v", client.listRequest)
	}
	if resp.GetTotal() != 1 || len(resp.GetList()) != 1 || resp.GetList()[0].GetId() != 1001 {
		t.Fatalf("response = %+v", resp)
	}
	if got := resp.GetList()[0]; got.GetPhone() != "138****8000" || got.GetIdCardNo() != "110101********1234" {
		t.Fatalf("masked user = %+v", got)
	}
}

// TestUserListCanUseUsersRPC_DisablesUnsupportedFilters 验证关键字筛选继续使用兼容路径。
func TestUserListCanUseUsersRPC_DisablesUnsupportedFilters(t *testing.T) {
	if userListCanUseUsersRPC(&adminsvc.UserListRequest{Keyword: "138"}) {
		t.Fatal("keyword filter should use compatibility query")
	}
}

// TestGetUser_UsesUsersRPC 验证用户详情优先查询真实 usersvc，并在 adminsvc 边界完成脱敏。
func TestGetUser_UsesUsersRPC(t *testing.T) {
	client := &fakeUsersClient{}
	logic := NewGetUserLogic(context.Background(), &svc.ServiceContext{UsersSvc: client})
	resp, err := logic.GetUser(&adminsvc.UserDetailRequest{Id: 1001})
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if client.detailRequest == nil || client.detailRequest.GetId() != 1001 {
		t.Fatalf("usersvc detail request = %+v", client.detailRequest)
	}
	if resp.GetNickname() != "乘客详情" || resp.GetPhone() != "138****8000" || resp.GetIdCardNo() != "110101********1234" {
		t.Fatalf("GetUser() response = %+v", resp)
	}
}

// TestGetUser_RevealsSensitiveForOpsRole 验证运营角色经过真实会话校验后可查看完整手机号和身份证号。
func TestGetUser_RevealsSensitiveForOpsRole(t *testing.T) {
	client := &fakeUsersClient{}
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.UsersSvc = client
	server := miniredis.RunT(t)
	defer server.Close()
	svcCtx.Redis = redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer svcCtx.Redis.Close()

	ctx := contextWithAdminSession(t, svcCtx, 2001, 2)
	mock.ExpectQuery(`SELECT id, username, password_hash, real_name, role, status\s+FROM admin_user\s+WHERE id = \? AND deleted_at IS NULL`).
		WithArgs(int64(2001)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "real_name", "role", "status"}).
			AddRow(2001, "ops", "hash", "运营", 2, 1))

	resp, err := NewGetUserLogic(ctx, svcCtx).GetUser(&adminsvc.UserDetailRequest{Id: 1001})
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if resp.GetPhone() != "13800138000" || resp.GetIdCardNo() != "110101199001011234" {
		t.Fatalf("GetUser() sensitive response = %+v", resp)
	}
}

// TestGetUser_MasksSensitiveForCustomerServiceRole 验证客服角色即使会话有效也只能查看脱敏用户信息。
func TestGetUser_MasksSensitiveForCustomerServiceRole(t *testing.T) {
	client := &fakeUsersClient{}
	svcCtx, mock, cleanup := newAdminSQLMock(t)
	defer cleanup()
	svcCtx.UsersSvc = client
	server := miniredis.RunT(t)
	defer server.Close()
	svcCtx.Redis = redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer svcCtx.Redis.Close()

	ctx := contextWithAdminSession(t, svcCtx, 3001, 3)
	mock.ExpectQuery(`SELECT id, username, password_hash, real_name, role, status\s+FROM admin_user\s+WHERE id = \? AND deleted_at IS NULL`).
		WithArgs(int64(3001)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "real_name", "role", "status"}).
			AddRow(3001, "cs", "hash", "客服", 3, 1))

	resp, err := NewGetUserLogic(ctx, svcCtx).GetUser(&adminsvc.UserDetailRequest{Id: 1001})
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if resp.GetPhone() != "138****8000" || resp.GetIdCardNo() != "110101********1234" {
		t.Fatalf("GetUser() masked response = %+v", resp)
	}
}

// TestMaskIDCard_ShortValueKeepsOriginal 验证异常短证件号不会被错误截断。
func TestMaskIDCard_ShortValueKeepsOriginal(t *testing.T) {
	if got := maskIDCard("123456"); got != "123456" {
		t.Fatalf("maskIDCard() = %q", got)
	}
}

// contextWithAdminSession 写入测试会话并构造携带 x-admin-token 的 gRPC 入站上下文。
// 该辅助函数模拟 api/admin 透传 token 到 adminsvc 的真实链路，不连接外部 Redis。
func contextWithAdminSession(t *testing.T, svcCtx *svc.ServiceContext, adminID int64, role int32) context.Context {
	t.Helper()
	token := "token-sensitive-test"
	if err := saveSession(context.Background(), svcCtx, adminSession{AdminID: adminID, Username: "admin", RealName: "管理员", Role: role, Status: 1, Token: token}); err != nil {
		t.Fatalf("saveSession() error = %v", err)
	}
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(adminTokenMetadataKey, token))
}
