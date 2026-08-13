package logic

import "XiaoLong-Ridy/rpc/usersvc/internal/model"

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = model.ErrInvalidPhone
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = model.ErrInvalidSMSCode
	// ErrSMSCodeExpired 表示验证码不存在或已过期。
	ErrSMSCodeExpired = model.ErrSMSCodeExpired
	// ErrSMSCodeSendTooFrequent 表示同手机号仍处于发送冷却期。
	ErrSMSCodeSendTooFrequent = model.ErrSMSCodeSendTooFrequent
	// ErrPhoneAlreadyExists 表示该手机号已注册。
	ErrPhoneAlreadyExists = model.ErrPhoneAlreadyExists
	// ErrUnsupportedRegister 表示注册来源不在支持范围内。
	ErrUnsupportedRegister = model.ErrUnsupportedRegister
	// ErrAccountFrozen 表示账号已被封禁。
	ErrAccountFrozen = model.ErrAccountFrozen
	// ErrInvalidToken 表示令牌格式、签名或状态无效。
	ErrInvalidToken = model.ErrInvalidToken
	// ErrTokenExpired 表示令牌已过期。
	ErrTokenExpired = model.ErrTokenExpired
)
