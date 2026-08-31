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

// UpdateProfileRequest 对应个人资料更新接口的请求参数。
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname,optional"`
	AvatarURL string `json:"avatarUrl,optional"`
}

// UpdateProfileResponse 对应个人资料更新接口的响应数据。
type UpdateProfileResponse struct {
	User UserInfo `json:"user"`
}

// SetPasswordRequest 对应已登录乘客设置或修改密码的请求参数。
type SetPasswordRequest struct {
	CurrentPassword string `json:"currentPassword,optional"`
	NewPassword     string `json:"newPassword"`
}

// SetPasswordResponse 返回密码设置结果，不包含密码或哈希。
type SetPasswordResponse struct {
	Success bool `json:"success"`
}
