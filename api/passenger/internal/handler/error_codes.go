package handler

const (
	// codeSuccess 表示请求处理成功。
	codeSuccess = 0
	// codeInvalidRequest 表示请求体或业务参数不合法。
	codeInvalidRequest = 40000
	// codeInvalidPhone 表示登录手机号格式不合法。
	codeInvalidPhone = 40001
	// codeTokenExpired 表示登录令牌已过期。
	codeTokenExpired = 40101
	// codeInvalidToken 表示登录令牌缺失、格式错误或签名无效。
	codeInvalidToken = 40102
	// codeAccountFrozen 表示乘客账号已被冻结。
	codeAccountFrozen = 40301
	// codeForbidden 表示当前乘客无权访问目标资源。
	codeForbidden = 40303
	// codeAddressNotFound 表示常用地址不存在或不属于当前乘客。
	codeAddressNotFound = 40401
	// codeUserNotFound 表示乘客账号不存在或不可用。
	codeUserNotFound = 40402
	// codeMethodNotAllowed 表示接口仅支持指定 HTTP 方法。
	codeMethodNotAllowed = 40500
	// codeInvalidSMSCode 表示短信验证码错误。
	codeInvalidSMSCode = 41001
	// codeSMSCodeExpired 表示短信验证码已过期。
	codeSMSCodeExpired = 41002
	// codeSMSCodeTooFrequent 表示短信验证码发送过于频繁。
	codeSMSCodeTooFrequent = 41003
	// codeInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	codeInvalidAddressPhone = 41011
	// codeInvalidLongitudeLatitude 表示常用地址或下单经纬度错误。
	codeInvalidLongitudeLatitude = 41014
	codeCouponNotFound           = 41015
	codeCouponUnavailable        = 41016
	codeCouponReceiveLimit       = 41017
	codeReviewAlreadyExists      = 41018
	// codeInternalError 表示服务器内部错误。
	codeInternalError = 50000
	// codeDownstreamUnavailable 表示下游 RPC 服务不可用。
	codeDownstreamUnavailable = 50001
)
