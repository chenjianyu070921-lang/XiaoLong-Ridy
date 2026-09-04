package proto

import "errors"

var (
	// ErrUserNotFound 表示用户不存在或已不可用。
	ErrUserNotFound = errors.New("user not found")
	// ErrAddressNotFound 表示常用地址不存在或不属于当前用户。
	ErrAddressNotFound = errors.New("address not found")
	// ErrInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	ErrInvalidAddressPhone = errors.New("address phone format invalid")
	// ErrInvalidLongitudeLatitude 表示常用地址经纬度为 0。
	ErrInvalidLongitudeLatitude = errors.New("longitude latitude invalid")
	// ErrCouponNotFound 表示优惠券模板不存在。
	ErrCouponNotFound = errors.New("coupon not found")
	// ErrCouponUnavailable 表示优惠券不可领取。
	ErrCouponUnavailable = errors.New("coupon unavailable")
	// ErrCouponReceiveLimit 表示用户领取次数已达上限。
	ErrCouponReceiveLimit = errors.New("coupon receive limit exceeded")
	// ErrUserCouponNotFound 表示用户优惠券不存在或不属于当前用户。
	ErrUserCouponNotFound = errors.New("user coupon not found")
	// ErrNicknameTooLong 表示昵称超过 20 字上限。
	ErrNicknameTooLong = errors.New("nickname too long, must be 20 characters or fewer")
)
