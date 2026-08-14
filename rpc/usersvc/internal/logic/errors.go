package logic

import "XiaoLong-Ridy/rpc/usersvc/internal/model"

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = model.ErrInvalidPhone
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = model.ErrInvalidSMSCode
	// ErrPhoneAlreadyExists 表示该手机号已注册。
	ErrPhoneAlreadyExists = model.ErrPhoneAlreadyExists
	// ErrUnsupportedRegister 表示注册来源不在支持范围内。
	ErrUnsupportedRegister = model.ErrUnsupportedRegister
	// ErrAccountFrozen 表示账号已被封禁。
	ErrAccountFrozen = model.ErrAccountFrozen
)
