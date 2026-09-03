package logic

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrCouponRepositoryNotConfigured 表示 usersvc 未注入优惠券仓储。
	ErrCouponRepositoryNotConfigured = errors.New("coupon repository not configured")
	// ErrInvalidCouponRequest 表示优惠券请求缺少必要参数。
	ErrInvalidCouponRequest = errors.New("invalid coupon request")
)

// ClaimCouponLogic 处理乘客领取优惠券 RPC。
type ClaimCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewClaimCouponLogic 创建领取优惠券逻辑实例。
func NewClaimCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClaimCouponLogic {
	return &ClaimCouponLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ClaimCoupon 校验请求并领取优惠券。
func (l *ClaimCouponLogic) ClaimCoupon(in *userproto.ClaimCouponRequest) (*userproto.ClaimCouponResponse, error) {
	if in.GetUserId() == 0 || in.GetCouponId() == 0 {
		return nil, ErrInvalidCouponRequest
	}
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	item, err := coupons.Claim(l.ctx, in.GetUserId(), in.GetCouponId())
	if err != nil {
		return nil, mapCouponRepositoryError(err)
	}
	return &userproto.ClaimCouponResponse{UserCoupon: toCouponInfo(item)}, nil
}

// ListMyCouponsLogic 处理查询我的优惠券 RPC。
type ListMyCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListMyCouponsLogic 创建我的优惠券列表逻辑实例。
func NewListMyCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyCouponsLogic {
	return &ListMyCouponsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListMyCoupons 查询指定乘客领取的优惠券列表，并在 usersvc 内部完成分页。
func (l *ListMyCouponsLogic) ListMyCoupons(in *userproto.ListMyCouponsRequest) (*userproto.ListMyCouponsResponse, error) {
	if in.GetUserId() == 0 || !isValidCouponStatus(in.GetStatus()) {
		return nil, ErrInvalidCouponRequest
	}
	page, pageSize := normalizeCouponPage(in.GetPage(), in.GetPageSize())
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	list, total, err := coupons.ListByUserPage(l.ctx, in.GetUserId(), int8(in.GetStatus()), page, pageSize)
	if err != nil {
		return nil, mapCouponRepositoryError(err)
	}
	resp := &userproto.ListMyCouponsResponse{List: make([]*userproto.CouponInfo, 0, len(list))}
	for _, item := range list {
		resp.List = append(resp.List, toCouponInfo(item))
	}
	resp.Total = total
	resp.Page = int32(page)
	resp.PageSize = int32(pageSize)
	return resp, nil
}

// AdminListUserCouponsLogic 处理管理后台用户券查询 RPC。
// 该逻辑支持按用户、券模板、状态和领取时间范围分页查询，供 adminsvc 进行运营检索。
type AdminListUserCouponsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAdminListUserCouponsLogic 创建后台用户券查询逻辑实例。
func NewAdminListUserCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListUserCouponsLogic {
	return &AdminListUserCouponsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AdminListUserCoupons 查询后台用户券列表。
func (l *AdminListUserCouponsLogic) AdminListUserCoupons(in *userproto.AdminListUserCouponsRequest) (*userproto.AdminListUserCouponsResponse, error) {
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizeCouponPage(in.GetPage(), in.GetPageSize())
	filter := repository.AdminUserCouponQuery{
		UserID:   in.GetUserId(),
		CouponID: in.GetCouponId(),
		Status:   int8(in.GetStatus()),
		Page:     page,
		PageSize: pageSize,
	}
	if strings.TrimSpace(in.GetStartTime()) != "" {
		filter.StartTime, err = time.ParseInLocation("2006-01-02 15:04:05", in.GetStartTime(), time.Local)
		if err != nil {
			return nil, ErrInvalidCouponRequest
		}
	}
	if strings.TrimSpace(in.GetEndTime()) != "" {
		filter.EndTime, err = time.ParseInLocation("2006-01-02 15:04:05", in.GetEndTime(), time.Local)
		if err != nil {
			return nil, ErrInvalidCouponRequest
		}
	}
	list, total, err := coupons.ListForAdminPage(l.ctx, filter)
	if err != nil {
		return nil, mapCouponRepositoryError(err)
	}
	resp := &userproto.AdminListUserCouponsResponse{List: make([]*userproto.CouponInfo, 0, len(list))}
	for _, item := range list {
		resp.List = append(resp.List, toCouponInfo(item))
	}
	resp.Total = total
	resp.Page = int32(page)
	resp.PageSize = int32(pageSize)
	return resp, nil
}

// AdminIssueCouponLogic 处理管理后台批量发券 RPC。
type AdminIssueCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAdminIssueCouponLogic 创建后台发券逻辑实例。
func NewAdminIssueCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminIssueCouponLogic {
	return &AdminIssueCouponLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// AdminIssueCoupon 批量向指定用户发放优惠券。
// 券模板库存、单用户领取上限与有效期由 usersvc 仓储统一校验，逐个用户独立事务，单个失败不影响其余用户。
func (l *AdminIssueCouponLogic) AdminIssueCoupon(in *userproto.AdminIssueCouponRequest) (*userproto.AdminIssueCouponResponse, error) {
	if in.GetCouponId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "优惠券ID不能为空")
	}
	if len(in.GetUserIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "发券目标用户不能为空")
	}
	if in.GetOperatorId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "操作管理员不能为空")
	}
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	var successCount, failCount int64
	failures := make([]string, 0)
	for _, userID := range in.GetUserIds() {
		if userID == 0 {
			failCount++
			failures = append(failures, "user_id=0: 无效用户")
			continue
		}
		if _, claimErr := coupons.Claim(l.ctx, userID, in.GetCouponId()); claimErr != nil {
			failCount++
			failures = append(failures, fmt.Sprintf("user_id=%d: %v", userID, mapCouponRepositoryError(claimErr)))
			continue
		}
		successCount++
	}
	return &userproto.AdminIssueCouponResponse{
		SuccessCount:   successCount,
		FailCount:      failCount,
		FailureReasons: failures,
	}, nil
}

// normalizeCouponPage 统一用户券查询的分页边界，避免下游收到无效 LIMIT/OFFSET。
func normalizeCouponPage(page, pageSize int32) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return int(page), int(pageSize)
}

// LockUserCouponLogic 处理下单前锁定用户券 RPC。
type LockUserCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLockUserCouponLogic 创建锁定用户券逻辑实例。
func NewLockUserCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockUserCouponLogic {
	return &LockUserCouponLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// LockUserCoupon 校验当前用户券状态、归属和适用范围后锁定优惠券。
func (l *LockUserCouponLogic) LockUserCoupon(in *userproto.LockUserCouponRequest) (*userproto.LockUserCouponResponse, error) {
	if in.GetUserId() == 0 || in.GetOrderId() == 0 {
		return nil, ErrInvalidCouponRequest
	}
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	item, err := coupons.Lock(l.ctx, in.GetUserId(), in.GetUserCouponId(), in.GetOrderId(), int8(in.GetCarType()), in.GetCityCode())
	if err != nil {
		return nil, mapCouponRepositoryError(err)
	}
	return &userproto.LockUserCouponResponse{Coupon: toCouponInfo(item)}, nil
}

// ReleaseUserCouponLogic 处理下单失败释放用户券 RPC。
type ReleaseUserCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewReleaseUserCouponLogic 创建释放用户券逻辑实例。
func NewReleaseUserCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseUserCouponLogic {
	return &ReleaseUserCouponLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ReleaseUserCoupon 将指定订单锁定的用户券释放回未使用状态。
func (l *ReleaseUserCouponLogic) ReleaseUserCoupon(in *userproto.ReleaseUserCouponRequest) (*userproto.ReleaseUserCouponResponse, error) {
	if in.GetUserId() == 0 || in.GetOrderId() == 0 {
		return nil, ErrInvalidCouponRequest
	}
	coupons, err := couponRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	var releaseErr error
	if in.GetUserCouponId() > 0 {
		releaseErr = coupons.Release(l.ctx, in.GetUserId(), in.GetUserCouponId(), in.GetOrderId())
	} else {
		releaseErr = coupons.ReleaseByOrder(l.ctx, in.GetUserId(), in.GetOrderId())
	}
	if releaseErr != nil {
		return nil, mapCouponRepositoryError(releaseErr)
	}
	return &userproto.ReleaseUserCouponResponse{Success: true}, nil
}

// ensureWelcomeCoupons 为历史账号补齐缺失的新人券，领取上限保证重复查询不会重复插入。
func ensureWelcomeCoupons(ctx context.Context, coupons repository.CouponRepository, userID uint64) error {
	for _, couponID := range []uint64{9001, 9002, 9003, 9004} {
		if _, err := coupons.Claim(ctx, userID, couponID); err != nil && !errors.Is(err, repository.ErrCouponReceiveLimit) {
			return err
		}
	}
	return nil
}

// couponRepository 获取 usersvc 优惠券仓储依赖。
func couponRepository(svcCtx *svc.ServiceContext) (repository.CouponRepository, error) {
	if svcCtx == nil || svcCtx.Coupons == nil {
		return nil, ErrCouponRepositoryNotConfigured
	}
	return svcCtx.Coupons, nil
}

// isValidCouponStatus 校验我的优惠券列表状态筛选；0 表示全部。
func isValidCouponStatus(status int32) bool {
	return status == 0 ||
		status == int32(model.UserCouponStatusUnused) ||
		status == int32(model.UserCouponStatusUsed) ||
		status == int32(model.UserCouponStatusExpired) ||
		status == int32(model.UserCouponStatusLocked)
}

// mapCouponRepositoryError 将仓储错误转换为 usersvc 对外错误。
func mapCouponRepositoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrCouponNotFound):
		return userproto.ErrCouponNotFound
	case errors.Is(err, repository.ErrCouponUnavailable):
		return userproto.ErrCouponUnavailable
	case errors.Is(err, repository.ErrCouponReceiveLimit):
		return userproto.ErrCouponReceiveLimit
	case errors.Is(err, repository.ErrUserCouponNotFound):
		return userproto.ErrUserCouponNotFound
	default:
		return err
	}
}

// toCouponInfo 将仓储聚合模型转换为 RPC 响应结构。
func toCouponInfo(item *repository.UserCouponWithTemplate) *userproto.CouponInfo {
	if item == nil || item.UserCoupon == nil || item.Coupon == nil {
		return nil
	}
	return &userproto.CouponInfo{
		UserCouponId:   item.UserCoupon.ID,
		CouponId:       item.Coupon.ID,
		Name:           item.Coupon.Name,
		Type:           int32(item.Coupon.Type),
		FaceValueCents: yuanToCents(item.Coupon.FaceValue),
		Discount:       discountPercent(item.Coupon.Discount),
		ThresholdCents: yuanToCents(item.Coupon.ThresholdAmount),
		CarType:        int32(item.Coupon.CarType),
		CityCode:       item.Coupon.CityCode,
		Status:         int32(item.UserCoupon.Status),
		ReceivedAt:     item.UserCoupon.ReceivedAt.Unix(),
		ExpireAt:       item.UserCoupon.ExpireAt.Unix(),
	}
}

// yuanToCents 将数据库中的元金额转换为接口使用的分。
func yuanToCents(value float64) int64 {
	return int64(math.Round(value * 100))
}

// discountPercent 将 0.85 这类折扣率转换为 85。
func discountPercent(value float64) int32 {
	return int32(math.Round(value * 100))
}
