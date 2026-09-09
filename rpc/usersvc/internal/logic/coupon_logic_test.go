package logic

import (
	"context"
	"testing"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// TestListMyCoupons_UsesRepositoryPagination 验证用户券列表在 usersvc 内部返回真实分页元数据，
// 避免上游服务必须拉取全量用户券后再做内存切片。
func TestListMyCoupons_UsesRepositoryPagination(t *testing.T) {
	repo := repository.NewMemoryCouponRepository()
	now := time.Now()
	for i := uint64(1); i <= 3; i++ {
		repo.AddCouponForTest(&model.Coupon{
			ID:           i,
			Name:         "测试优惠券",
			Type:         model.CouponTypeInstantReduction,
			ValidStartAt: now.Add(-time.Hour),
			ValidEndAt:   now.Add(time.Hour),
			Status:       model.CouponStatusEnabled,
		})
		if _, err := repo.Claim(context.Background(), 1001, i); err != nil {
			t.Fatalf("Claim() coupon %d error = %v", i, err)
		}
	}

	logic := NewListMyCouponsLogic(context.Background(), &svc.ServiceContext{Coupons: repo})
	resp, err := logic.ListMyCoupons(&userproto.ListMyCouponsRequest{UserId: 1001, Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("ListMyCoupons() error = %v", err)
	}
	if resp.GetTotal() != 3 || resp.GetPage() != 2 || resp.GetPageSize() != 2 {
		t.Fatalf("ListMyCoupons() page meta = total:%d page:%d size:%d, want 3/2/2", resp.GetTotal(), resp.GetPage(), resp.GetPageSize())
	}
	if len(resp.GetList()) != 1 {
		t.Fatalf("ListMyCoupons() list len = %d, want 1", len(resp.GetList()))
	}
}

// TestNormalizeCouponPage 验证用户券分页边界，避免无效分页传入仓储层。
func TestNormalizeCouponPage(t *testing.T) {
	page, size := normalizeCouponPage(0, 1000)
	if page != 1 || size != 100 {
		t.Fatalf("normalizeCouponPage() = %d/%d, want 1/100", page, size)
	}
}
