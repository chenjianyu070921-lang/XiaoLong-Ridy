package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

var errInvalidAvatarImage = errors.New("avatar image data is invalid")

var ErrQiniuNotConfigured = errors.New("qiniu storage not configured")

type AvatarLogic struct{}

func NewAvatarLogic() *AvatarLogic {
	return &AvatarLogic{}
}

func (l *AvatarLogic) GetUploadToken(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, req *types.AvatarUploadTokenRequest) (*types.AvatarUploadTokenResponse, error) {
	if driverID <= 0 {
		return nil, errors.New("driver id is invalid")
	}
	if req == nil || strings.TrimSpace(req.Extension) == "" {
		return nil, errInvalidAvatarImage
	}
	if svcCtx == nil || svcCtx.Qiniu == nil {
		return nil, ErrQiniuNotConfigured
	}
	token, err := svcCtx.Qiniu.GenerateAvatarToken(ctx, uint64(driverID), req.Extension)
	if err != nil {
		return nil, errInvalidAvatarImage
	}
	return &types.AvatarUploadTokenResponse{
		UploadToken: token.UploadToken,
		Key:         token.Key,
		Domain:      token.Domain,
		UploadURL:   token.UploadURL,
		ExpireSec:   token.ExpireSec,
	}, nil
}
