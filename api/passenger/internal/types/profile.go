package types

// GetProfileRequest 对应个人中心资料查询接口的请求参数。
type GetProfileRequest struct{}

// GetProfileResponse 对应个人中心资料查询接口的响应数据。
type GetProfileResponse struct {
	User UserInfo `json:"user"`
}

// SubmitRealNameRequest 对应实名资料提交接口的请求参数。
type SubmitRealNameRequest struct {
	RealName string `json:"realName"`
	IDCardNo string `json:"idCardNo"`
}

// SubmitRealNameResponse 对应实名资料提交接口的响应数据。
type SubmitRealNameResponse struct {
	User UserInfo `json:"user"`
}
