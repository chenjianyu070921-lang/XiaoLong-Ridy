package logic

import (
	"errors"

	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

var (
	// ErrInvalidPhone 表示手机号格式不合法。
	ErrInvalidPhone = userproto.ErrInvalidPhone
	// ErrInvalidSMSCode 表示短信验证码不正确。
	ErrInvalidSMSCode = userproto.ErrInvalidSMSCode
	// ErrPhoneAlreadyExists 表示该手机号已注册。
	ErrPhoneAlreadyExists = errors.New("phone already registered")
	// ErrUnsupportedRegister 表示注册来源不在支持范围内。
	ErrUnsupportedRegister = errors.New("unsupported register source")
	// ErrAccountFrozen 表示账号已被封禁。
	ErrAccountFrozen = userproto.ErrAccountFrozen
)
