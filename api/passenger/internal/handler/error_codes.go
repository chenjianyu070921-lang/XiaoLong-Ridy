package handler

const (
	// codeSuccess 表示请求处理成功。
	codeSuccess = 0
	// codeInvalidRequest 表示请求参数或业务前置校验不合法。
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
	// codeMethodNotAllowed 表示接口只支持指定 HTTP 方法。
	codeMethodNotAllowed = 40500
	// codeInvalidSMSCode 表示短信验证码错误。
	codeInvalidSMSCode = 41001
	// codeSMSCodeExpired 表示短信验证码已过期。
	codeSMSCodeExpired = 41002
	// codeSMSCodeTooFrequent 表示短信验证码发送过于频繁。
	codeSMSCodeTooFrequent = 41003
	// codeInvalidAddressPhone 表示常用地址联系人手机号格式错误。
	codeInvalidAddressPhone = 41011
	// codeInvalidLongitudeLatitude 表示常用地址经纬度错误。
	codeInvalidLongitudeLatitude = 41014
	// codeCouponNotFound 表示优惠券或用户券不存在。
	codeCouponNotFound = 41021
	// codeCouponUnavailable 表示优惠券当前不可用。
	codeCouponUnavailable = 41022
	// codeCouponReceiveLimit 表示优惠券领取次数已达上限。
	codeCouponReceiveLimit = 41023
	// codeReviewAlreadyExists 表示订单已评价，不能重复提交评价。
	codeReviewAlreadyExists = 41031
	// codeRealNameVerifyFailed 表示实名认证未通过或第三方核验调用失败。
	codeRealNameVerifyFailed = 42001
	// codeInternalError 表示服务内部错误。
	codeInternalError = 50000
	// codeDownstreamUnavailable 表示下游 RPC 服务不可用。
	codeDownstreamUnavailable = 50001
)
