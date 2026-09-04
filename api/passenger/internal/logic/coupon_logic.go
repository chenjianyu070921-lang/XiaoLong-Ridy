package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// CouponLogic 封装乘客端优惠券领取和查询流程。
type CouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewCouponLogic 创建优惠券业务逻辑实例。
func NewCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *CouponLogic {
	return &CouponLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// ClaimCoupon 领取指定优惠券模板。
func (l *CouponLogic) ClaimCoupon(req *types.ClaimCouponRequest) (*types.ClaimCouponResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.CouponID == 0 {
		return nil, ErrInvalidRequest
	}
	userClient, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := userClient.ClaimCoupon(l.ctx, &userproto.ClaimCouponRequest{
		UserId:   userID,
		CouponId: req.CouponID,
	})
	if err != nil {
		return nil, err
	}
	return &types.ClaimCouponResponse{UserCoupon: toPassengerCouponInfo(resp.GetUserCoupon())}, nil
}

// ClaimWelcomeGift 仅允许没有任何用户券的新用户领取一次新人礼包。
func (l *CouponLogic) ClaimWelcomeGift() (*types.ClaimWelcomeGiftResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	existing, err := client.ListMyCoupons(l.ctx, &userproto.ListMyCouponsRequest{UserId: userID, Status: 0})
	if err != nil {
		return nil, err
	}
	if len(existing.GetList()) > 0 {
		return nil, errors.New("新人礼包仅限首次登录用户领取")
	}
	ids := []uint64{9001, 9002, 9003, 9004}
	result := make([]types.CouponInfo, 0, len(ids))
	for _, id := range ids {
		item, claimErr := client.ClaimCoupon(l.ctx, &userproto.ClaimCouponRequest{UserId: userID, CouponId: id})
		if claimErr != nil {
			return nil, claimErr
		}
		result = append(result, toPassengerCouponInfo(item.GetUserCoupon()))
	}
	return &types.ClaimWelcomeGiftResponse{List: result}, nil
}

// ListMyCoupons 查询当前乘客自己的优惠券列表。
func (l *CouponLogic) ListMyCoupons(req *types.ListMyCouponsRequest) (*types.ListMyCouponsResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || !isValidUserCouponStatus(req.Status) {
		return nil, ErrInvalidRequest
	}
	userClient, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := userClient.ListMyCoupons(l.ctx, &userproto.ListMyCouponsRequest{
		UserId: userID,
		Status: req.Status,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.CouponInfo, 0, len(resp.GetList()))
	for _, item := range resp.GetList() {
		list = append(list, toPassengerCouponInfo(item))
	}
	return &types.ListMyCouponsResponse{List: list}, nil
}

// userClient 获取用户服务客户端，优惠券归属 usersvc 统一处理。
func (l *CouponLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}

// isValidUserCouponStatus 校验用户券状态筛选，0 表示全部。
func isValidUserCouponStatus(status int32) bool {
	return status >= 0 && status <= 4
}

// toPassengerCouponInfo 将 usersvc 优惠券响应转换为 passenger HTTP 响应。
func toPassengerCouponInfo(item *userproto.CouponInfo) types.CouponInfo {
	if item == nil {
		return types.CouponInfo{}
	}
	return types.CouponInfo{
		UserCouponID:   item.GetUserCouponId(),
		CouponID:       item.GetCouponId(),
		Name:           item.GetName(),
		Type:           item.GetType(),
		FaceValueCents: item.GetFaceValueCents(),
		Discount:       item.GetDiscount(),
		ThresholdCents: item.GetThresholdCents(),
		CarType:        item.GetCarType(),
		CityCode:       item.GetCityCode(),
		Status:         item.GetStatus(),
		ReceivedAt:     item.GetReceivedAt(),
		ExpireAt:       item.GetExpireAt(),
	}
}
