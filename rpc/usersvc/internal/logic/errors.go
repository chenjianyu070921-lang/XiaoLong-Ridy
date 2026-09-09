package logic

import "XiaoLong-Ridy/rpc/usersvc/internal/model"

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = model.ErrInvalidPhone
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = model.ErrInvalidSMSCode
	// ErrSMSCodeExpired 表示短信验证码不存在或已经过期。
	ErrSMSCodeExpired = model.ErrSMSCodeExpired
	// ErrSMSCodeSendTooFrequent 表示同一手机号发送验证码过于频繁。
	ErrSMSCodeSendTooFrequent = model.ErrSMSCodeSendTooFrequent
	// ErrPhoneAlreadyExists 表示该手机号已注册。
	ErrPhoneAlreadyExists = model.ErrPhoneAlreadyExists
	// ErrUnsupportedRegister 表示注册来源不在支持范围内。
	ErrUnsupportedRegister = model.ErrUnsupportedRegister
	// ErrAccountFrozen 表示账号已被封禁。
	ErrAccountFrozen = model.ErrAccountFrozen
	// ErrInvalidToken 表示令牌格式、签名或服务端状态无效。
	ErrInvalidToken = model.ErrInvalidToken
	// ErrTokenExpired 表示令牌已经超过有效期。
	ErrTokenExpired = model.ErrTokenExpired
	// ErrInvalidRealNameInfo 表示实名资料缺少真实姓名或证件号。
	ErrInvalidRealNameInfo = model.ErrInvalidRealNameInfo
	// ErrRealNameVerifyFailed 表示腾讯云实名认证校验未通过或调用失败。
	ErrRealNameVerifyFailed = model.ErrRealNameVerifyFailed
)
