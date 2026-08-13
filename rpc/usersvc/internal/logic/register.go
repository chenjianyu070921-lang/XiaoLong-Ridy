package logic

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
)

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

type SMSCodeVerifier interface {
	// Verify 校验手机号和验证码是否匹配。
	Verify(ctx context.Context, phone, code string) (bool, error)
}

type RegisterRequest struct {
	Phone          string
	SMSCode        string
	Password       string
	Nickname       string
	RegisterSource string
}

// RegisterResponse 返回注册成功后所需的最小身份信息。
type RegisterResponse struct {
	UserID         uint64
	Phone          string
	Nickname       string
	RegisterSource string
}

type RegisterLogic struct {
	users    repository.UserRepository
	verifier SMSCodeVerifier
}

// NewRegisterLogic 组装注册场景所需的仓储和验证码校验依赖。
func NewRegisterLogic(users repository.UserRepository, verifier SMSCodeVerifier) *RegisterLogic {
	return &RegisterLogic{
		users:    users,
		verifier: verifier,
	}
}

// Register 负责校验输入、验码、查重并创建账号。
func (l *RegisterLogic) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	if !phonePattern.MatchString(phone) {
		return nil, ErrInvalidPhone
	}

	source := normalizeRegisterSource(req.RegisterSource)
	if source == "" {
		return nil, ErrUnsupportedRegister
	}

	ok, err := l.verifier.Verify(ctx, phone, strings.TrimSpace(req.SMSCode))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidSMSCode
	}

	if _, err := l.users.FindByPhone(ctx, phone); err == nil {
		return nil, ErrPhoneAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	passwordHash := ""
	if req.Password != "" {
		passwordHash, err = hashPassword(req.Password)
		if err != nil {
			return nil, err
		}
	}

	user := &model.User{
		Phone:          phone,
		PasswordHash:   passwordHash,
		Nickname:       strings.TrimSpace(req.Nickname),
		RegisterSource: source,
		Status:         model.UserStatusNormal,
	}
	if user.Nickname == "" {
		user.Nickname = maskPhone(phone)
	}

	if err := l.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrPhoneExists) {
			return nil, ErrPhoneAlreadyExists
		}
		return nil, err
	}

	return &RegisterResponse{
		UserID:         user.ID,
		Phone:          user.Phone,
		Nickname:       user.Nickname,
		RegisterSource: user.RegisterSource,
	}, nil
}

// normalizeRegisterSource 将注册来源规整到支持的范围内。
func normalizeRegisterSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return model.RegisterSourcePhone
	}

	switch source {
	case model.RegisterSourcePhone, model.RegisterSourceWechat, model.RegisterSourceAlipay:
		return source
	default:
		return ""
	}
}

// maskPhone 根据手机号生成默认展示昵称。
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// hashPassword 将密码转换为带盐哈希，便于后续密码登录扩展。
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	sum := sha256.Sum256(append(salt, []byte(password)...))
	return fmt.Sprintf("sha256$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum[:]),
	), nil
}
