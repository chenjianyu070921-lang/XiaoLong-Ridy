package logic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/types"
)

var errInvalidAvatarImage = errors.New("avatar image data is invalid")

const maxAvatarImageBytes = 5 * 1024 * 1024

type AvatarLogic struct{}

func NewAvatarLogic() *AvatarLogic {
	return &AvatarLogic{}
}

func (l *AvatarLogic) UploadDriverAvatar(driverID int64, req *types.UploadDriverAvatarRequest) (*types.UploadDriverAvatarResponse, error) {
	if driverID <= 0 {
		return nil, errors.New("driver id is invalid")
	}
	if req == nil || strings.TrimSpace(req.Avatar) == "" {
		return nil, errInvalidAvatarImage
	}
	data, err := decodeAvatarImage(req.Avatar)
	if err != nil {
		return nil, err
	}
	ext := guessAvatarImageExt(req.Avatar)
	objectKey := fmt.Sprintf("drivers/%d/avatar-%d%s", driverID, time.Now().UnixNano(), ext)
	target := filepath.Join(LocalAvatarDir(), filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(target, data, 0o600); err != nil {
		return nil, err
	}
	return &types.UploadDriverAvatarResponse{
		AvatarURL: strings.TrimRight(LocalAvatarPublicPrefix(), "/") + "/" + objectKey,
	}, nil
}

func decodeAvatarImage(input string) ([]byte, error) {
	input = strings.TrimSpace(input)
	if idx := strings.Index(input, ","); strings.HasPrefix(input, "data:") && idx >= 0 {
		input = input[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, errInvalidAvatarImage
	}
	if len(data) == 0 || len(data) > maxAvatarImageBytes {
		return nil, errInvalidAvatarImage
	}
	return data, nil
}

func guessAvatarImageExt(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	if strings.HasPrefix(lower, "data:image/png") {
		return ".png"
	}
	if strings.HasPrefix(lower, "data:image/webp") {
		return ".webp"
	}
	if strings.HasPrefix(lower, "data:image/jpeg") || strings.HasPrefix(lower, "data:image/jpg") {
		return ".jpg"
	}
	head := lower
	if len(head) > 8 {
		head = head[:8]
	}
	if strings.HasPrefix(head, "ivbor") {
		return ".png"
	}
	if strings.HasPrefix(head, "uklgr") {
		return ".webp"
	}
	return ".jpg"
}

func LocalAvatarDir() string {
	if dir := strings.TrimSpace(os.Getenv("DRIVER_AVATAR_LOCAL_DIR")); dir != "" {
		return dir
	}
	return filepath.Join(".run", "avatars")
}

func LocalAvatarPublicPrefix() string {
	if prefix := strings.TrimSpace(os.Getenv("DRIVER_AVATAR_PUBLIC_PREFIX")); prefix != "" {
		return prefix
	}
	return "/api/driver/v1/avatar-files"
}
