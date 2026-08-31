package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
	"golang.org/x/crypto/bcrypt"
)

// LoginByPasswordLogic 负责校验手机号密码并签发登录令牌。
type LoginByPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginByPasswordLogic 创建密码登录业务对象。
func NewLoginByPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByPasswordLogic {
	return &LoginByPasswordLogic{ctx: ctx, svcCtx: svcCtx}
}

// LoginByPassword 查询账号、校验 bcrypt 密码并返回标准登录响应。
func (l *LoginByPasswordLogic) LoginByPassword(in *userproto.LoginByPasswordRequest) (*userproto.LoginBySMSResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	if !IsValidPhone(phone) {
		return nil, ErrInvalidPhone
	}
	if strings.TrimSpace(in.GetPassword()) == "" {
		return nil, errors.New("invalid password")
	}
	users, err := userRepository(l.svcCtx)
	if err != nil {
		return nil, err
	}
	user, err := users.FindByPhone(l.ctx, phone)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if user.Status == model.UserStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if user.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.GetPassword())) != nil {
		return nil, errors.New("invalid credentials")
	}
	token, refresh, err := l.svcCtx.Tokens.Issue(user.ID, user.Phone, user.Status)
	if err != nil {
		return nil, err
	}
	return &userproto.LoginBySMSResponse{Token: token, RefreshToken: refresh, User: toUserInfo(user)}, nil
}
