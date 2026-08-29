package logic

import (
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	"context"
	"strings"
)

// AvatarUploadLogic 处理乘客头像上传凭证申请，不接收或转发图片二进制内容。
type AvatarUploadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewAvatarUploadLogic 创建头像上传凭证逻辑实例。
func NewAvatarUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *AvatarUploadLogic {
	return &AvatarUploadLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// GetToken 校验登录态和扩展名后生成七牛云凭证。
func (l *AvatarUploadLogic) GetToken(req *types.AvatarUploadTokenRequest) (*types.AvatarUploadTokenResponse, error) {
	uid, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || strings.TrimSpace(req.Extension) == "" {
		return nil, ErrInvalidRequest
	}
	if l.svcCtx.Qiniu == nil {
		return nil, ErrQiniuNotConfigured
	}
	r, err := l.svcCtx.Qiniu.GenerateAvatarToken(l.ctx, uid, req.Extension)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	return &types.AvatarUploadTokenResponse{UploadToken: r.UploadToken, Key: r.Key, Domain: r.Domain, UploadURL: r.UploadURL, ExpireSec: r.ExpireSec}, nil
}

var ErrQiniuNotConfigured = errQiniuNotConfigured{}

type errQiniuNotConfigured struct{}

func (errQiniuNotConfigured) Error() string { return "qiniu client not configured" }
