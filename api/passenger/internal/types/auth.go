package types

// SendSMSCodeRequest 对应发送短信验证码接口的请求参数。
type SendSMSCodeRequest struct {
	Phone string `json:"phone"`
}

// SendSMSCodeResponse 对应发送短信验证码接口的响应数据。
type SendSMSCodeResponse struct {
	Success  bool  `json:"success"`
	ExpireIn int64 `json:"expireIn"`
}

// LoginBySMSRequest 对应短信验证码登录接口的请求参数。
type LoginBySMSRequest struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// UserInfo 是登录接口返回的用户基础信息。
type UserInfo struct {
	UserID         uint64 `json:"userId"`
	Phone          string `json:"phone"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatarUrl"`
	RealNameStatus string `json:"realNameStatus"`
}

// LoginBySMSResponse 对应短信验证码登录接口的响应数据。
type LoginBySMSResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	IsNewUser    bool     `json:"isNewUser"`
	User         UserInfo `json:"user"`
}

// RefreshTokenRequest 对应刷新令牌接口的请求参数。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenResponse 对应刷新令牌接口的响应数据。
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

// LogoutResponse 对应登出接口的响应数据。
type LogoutResponse struct {
	Success bool `json:"success"`
}

// Response 是接口文档约定的通用响应格式。
type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Timestamp int64  `json:"timestamp"`
	TraceID   string `json:"traceId"`
}
