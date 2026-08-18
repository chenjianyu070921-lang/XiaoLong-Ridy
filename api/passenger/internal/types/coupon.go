package types

// ClaimCouponRequest 表示乘客领取优惠券的请求参数。
type ClaimCouponRequest struct {
	CouponID uint64 `json:"couponId"`
}

// ClaimCouponResponse 表示领取优惠券后的响应数据。
type ClaimCouponResponse struct {
	UserCoupon CouponInfo `json:"userCoupon"`
}

// ListMyCouponsRequest 表示我的优惠券列表查询请求参数。
type ListMyCouponsRequest struct {
	Status int32 `json:"status"`
}

// ListMyCouponsResponse 表示我的优惠券列表响应数据。
type ListMyCouponsResponse struct {
	List []CouponInfo `json:"list"`
}

// CouponInfo 表示乘客端展示的优惠券信息。
type CouponInfo struct {
	UserCouponID   uint64 `json:"userCouponId"`
	CouponID       uint64 `json:"couponId"`
	Name           string `json:"name"`
	Type           int32  `json:"type"`
	FaceValueCents int64  `json:"faceValueCents"`
	Discount       int32  `json:"discount"`
	ThresholdCents int64  `json:"thresholdCents"`
	CarType        int32  `json:"carType"`
	CityCode       string `json:"cityCode"`
	Status         int32  `json:"status"`
	ReceivedAt     int64  `json:"receivedAt"`
	ExpireAt       int64  `json:"expireAt"`
}
