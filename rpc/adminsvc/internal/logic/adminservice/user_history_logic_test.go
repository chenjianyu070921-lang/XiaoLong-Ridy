package adminservicelogic

import (
	"context"
	"testing"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"
	usersvc "XiaoLong-Ridy/rpc/usersvc/proto"

	"google.golang.org/grpc"
)

// fakeUserCouponClient 只覆盖本测试关注的用户券分页 RPC。
// 通过匿名嵌入 usersvc.UserClient 保持接口兼容，未调用的方法不会参与测试。
type fakeUserCouponClient struct {
	usersvc.UserClient
	got *usersvc.ListMyCouponsRequest
}

func (f *fakeUserCouponClient) ListMyCoupons(ctx context.Context, in *usersvc.ListMyCouponsRequest, opts ...grpc.CallOption) (*usersvc.ListMyCouponsResponse, error) {
	f.got = in
	return &usersvc.ListMyCouponsResponse{
		List:     []*usersvc.CouponInfo{{UserCouponId: 11, CouponId: 22, Name: "测试券"}},
		Total:    88,
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
	}, nil
}

// TestListUserCoupons_ForwardsPaginationToUserSvc 验证 adminsvc 不再拉取全量用户券后内存分页，
// 而是将分页参数透传给 usersvc，由下游数据库查询完成分页和总数统计。
func TestListUserCoupons_ForwardsPaginationToUserSvc(t *testing.T) {
	client := &fakeUserCouponClient{}
	logic := NewUserHistoryLogic(context.Background(), &svc.ServiceContext{UsersSvc: client})

	resp, err := logic.ListUserCoupons(&adminsvc.UserCouponHistoryRequest{UserId: 1001, Status: 1, Page: 3, PageSize: 15})
	if err != nil {
		t.Fatalf("ListUserCoupons() error = %v", err)
	}
	if client.got == nil || client.got.GetUserId() != 1001 || client.got.GetStatus() != 1 || client.got.GetPage() != 3 || client.got.GetPageSize() != 15 {
		t.Fatalf("ListUserCoupons() forwarded request = %+v, want user/status/page/page_size", client.got)
	}
	if resp.GetTotal() != 88 || resp.GetPage() != 3 || resp.GetPageSize() != 15 || len(resp.GetList()) != 1 {
		t.Fatalf("ListUserCoupons() response = %+v, want downstream pagination result", resp)
	}
}
