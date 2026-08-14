package proto

import "errors"

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = errors.New("invalid phone")
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = errors.New("invalid sms code")
	// ErrSMSCodeExpired 表示验证码不存在或已过期。
	ErrSMSCodeExpired = errors.New("sms code expired")
	// ErrSMSCodeSendTooFrequent 表示同手机号仍处于发送冷却期。
	ErrSMSCodeSendTooFrequent = errors.New("sms code send too frequent")
	// ErrAccountFrozen 表示账号已被封禁。
	ErrAccountFrozen = errors.New("account frozen")
	// ErrInvalidToken 表示令牌格式、签名或状态无效。
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired 表示令牌已过期。
	ErrTokenExpired = errors.New("token expired")
	// ErrAddressNotFound 表示常用地址不存在或不属于当前用户。
	ErrAddressNotFound = errors.New("address not found")
	// ErrInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	ErrInvalidAddressPhone = errors.New("异常_电话格式错误")
	// ErrInvalidLongitudeLatitude 表示常用地址经纬度为 0。
	ErrInvalidLongitudeLatitude = errors.New("异常_经纬度为0")
)
