package logic

import "errors"

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = errors.New("invalid phone")
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = errors.New("invalid sms code")
	// ErrPhoneAlreadyExists 表示该手机号已注册。
	ErrPhoneAlreadyExists = errors.New("phone already registered")
	// ErrUnsupportedRegister 表示注册来源不在支持范围内。
	ErrUnsupportedRegister = errors.New("unsupported register source")
)
