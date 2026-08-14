package logic

import userproto "XiaoLong-Ridy/rpc/usersvc/proto"

var (
	// ErrSMSCodeExpired 表示验证码不存在或已过期。
	ErrSMSCodeExpired = userproto.ErrSMSCodeExpired
	// ErrSMSCodeSendTooFrequent 表示同手机号仍处于发送冷却期。
	ErrSMSCodeSendTooFrequent = userproto.ErrSMSCodeSendTooFrequent
	// ErrInvalidToken 表示令牌格式、签名或状态无效。
	ErrInvalidToken = userproto.ErrInvalidToken
	// ErrTokenExpired 表示令牌已过期。
	ErrTokenExpired = userproto.ErrTokenExpired
)
