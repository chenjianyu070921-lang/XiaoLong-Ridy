package logic

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidPassword 表示密码未满足长度与复杂度要求。
	ErrInvalidPassword = errors.New("invalid password")
	// ErrCurrentPasswordIncorrect 表示已设置密码的账号未通过旧密码校验。
	ErrCurrentPasswordIncorrect = errors.New("current password incorrect")
)

// SetPasswordLogic 处理已登录乘客设置或修改密码的流程。
type SetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSetPasswordLogic 创建密码设置逻辑对象。
func NewSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPasswordLogic {
	return &SetPasswordLogic{ctx: ctx, svcCtx: svcCtx}
}

// SetPassword 使用 bcrypt 保存密码哈希；账户已有密码时需先验证旧密码。
func (l *SetPasswordLogic) SetPassword(in *userproto.SetPasswordRequest) (*userproto.SetPasswordResponse, error) {
	if in.GetUserId() == 0 || !validPassword(in.GetNewPassword()) {
		return nil, ErrInvalidPassword
	}
	users, err := userRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	user, err := users.FindByID(l.ctx, in.GetUserId())
	if err != nil {
		return nil, mapUserRepositoryError(err)
	}
	if user.Status == model.UserStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if user.PasswordHash != "" && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.GetCurrentPassword())) != nil {
		return nil, ErrCurrentPasswordIncorrect
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hash)
	if err := users.Update(l.ctx, user); err != nil {
		return nil, mapUserRepositoryError(err)
	}
	return &userproto.SetPasswordResponse{Success: true}, nil
}

// validPassword 要求密码长度为 8-64 个字符并同时包含字母与数字。
func validPassword(password string) bool {
	password = strings.TrimSpace(password)
	if length := len([]rune(password)); length < 8 || length > 32 {
		return false
	}
	var hasLetter, hasDigit bool
	for _, char := range password {
		hasLetter = hasLetter || unicode.IsLetter(char)
		hasDigit = hasDigit || unicode.IsDigit(char)
	}
	return hasLetter && hasDigit
}
