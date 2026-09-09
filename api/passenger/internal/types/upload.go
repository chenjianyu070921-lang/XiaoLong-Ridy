package types

// AvatarUploadTokenRequest 描述获取头像上传凭证时提交的文件扩展名。
type AvatarUploadTokenRequest struct {
	Extension string `json:"extension"`
}

// AvatarUploadTokenResponse 返回前端直传七牛云所需参数。
type AvatarUploadTokenResponse struct {
	UploadToken string `json:"uploadToken"`
	Key         string `json:"key"`
	Domain      string `json:"domain"`
	UploadURL   string `json:"uploadUrl"`
	ExpireSec   int64  `json:"expireSec"`
}
