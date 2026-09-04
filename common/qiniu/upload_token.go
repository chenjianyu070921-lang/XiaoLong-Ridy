// Package qiniu 封装七牛云对象存储的上传凭证生成能力。
package qiniu

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/uptoken"
)

// Config 保存七牛云客户端运行配置，密钥应通过环境变量注入。
type Config struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
	UploadURL string
	ExpireSec int64
}

// Client 负责按照固定业务前缀生成七牛云客户端上传凭证。
type Client struct{ cfg Config }

// NewClient 创建七牛云凭证客户端，并校验必填配置。
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" || strings.TrimSpace(cfg.Bucket) == "" || strings.TrimSpace(cfg.Domain) == "" {
		return nil, fmt.Errorf("qiniu access key, secret key, bucket and domain are required")
	}
	if cfg.ExpireSec <= 0 {
		cfg.ExpireSec = 600
	}
	if strings.TrimSpace(cfg.UploadURL) == "" {
		cfg.UploadURL = "https://upload.qiniup.com"
	}
	return &Client{cfg: cfg}, nil
}

// UploadTokenResponse 是前端直传七牛云所需的完整参数。
type UploadTokenResponse struct {
	UploadToken, Key, Domain, UploadURL string
	ExpireSec                           int64
}

// GenerateAvatarToken 生成乘客头像的唯一对象 key 和简单上传凭证。
func (c *Client) GenerateAvatarToken(ctx context.Context, userID uint64, extension string) (*UploadTokenResponse, error) {
	ext := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(extension), "."))
	if ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "webp" {
		return nil, fmt.Errorf("unsupported avatar extension")
	}
	key := fmt.Sprintf("avatar/passenger/%d/%d.%s", userID, time.Now().UnixNano(), ext)
	policy, err := uptoken.NewPutPolicy(c.cfg.Bucket+":"+key, time.Now().Add(time.Duration(c.cfg.ExpireSec)*time.Second))
	if err != nil {
		return nil, err
	}
	token, err := uptoken.NewSigner(policy, credentials.NewCredentials(c.cfg.AccessKey, c.cfg.SecretKey)).GetUpToken(ctx)
	if err != nil {
		return nil, err
	}
	return &UploadTokenResponse{UploadToken: token, Key: filepath.ToSlash(key), Domain: strings.TrimRight(c.cfg.Domain, "/"), UploadURL: c.cfg.UploadURL, ExpireSec: c.cfg.ExpireSec}, nil
}
