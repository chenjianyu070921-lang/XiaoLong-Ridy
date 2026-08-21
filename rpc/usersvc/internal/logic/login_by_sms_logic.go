package logic

import (
	"context"
	"errors"
	"strings"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/repository"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

// LoginBySMSLogic 处理短信验证码登录 RPC 的完整业务流程。
type LoginBySMSLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLoginBySMSLogic 创建短信登录逻辑实例，并绑定请求上下文与服务依赖。
func NewLoginBySMSLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginBySMSLogic {
	return &LoginBySMSLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// LoginBySMS 校验验证码并完成登录；首次手机号登录时自动创建乘客账号。
func (l *LoginBySMSLogic) LoginBySMS(in *userproto.LoginBySMSRequest) (*userproto.LoginBySMSResponse, error) {
	phone := strings.TrimSpace(in.GetPhone())
	if !IsValidPhone(phone) {
		return nil, ErrInvalidPhone
	}

	// 先校验验证码，再进行用户查询或自动注册，避免无效请求污染用户数据。
	ok, err := l.svcCtx.SMSVerifier.Verify(l.ctx, phone, strings.TrimSpace(in.GetCode()))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidSMSCode
	}

	user, isNewUser, err := l.findOrCreateUser(phone)
	if err != nil {
		return nil, err
	}
	if user.Status == model.UserStatusFrozen {
		return nil, ErrAccountFrozen
	}
	if err := l.recordBlacklistLoginHit(user.ID); err != nil {
		return nil, err
	}

	// 登录成功后统一签发 Access Token 和 Refresh Token，供乘客端后续接口鉴权使用。
	token, refreshToken, err := l.svcCtx.Tokens.Issue(user.ID, user.Phone, user.Status)
	if err != nil {
		return nil, err
	}

	return &userproto.LoginBySMSResponse{
		Token:        token,
		RefreshToken: refreshToken,
		IsNewUser:    isNewUser,
		User:         toUserInfo(user),
	}, nil
}

// recordBlacklistLoginHit 检查登录用户是否命中生效黑名单，并落库保存登录场景审计记录。
// 基础表仅支持数值 target_id，因此登录按已定位的用户 ID 查询，而不将手机号错误转换为数值 ID。
// 查询基础设施异常不阻断登录；已确认命中后记录写入失败必须返回错误，避免命中事实静默丢失。
func (l *LoginBySMSLogic) recordBlacklistLoginHit(userID uint64) error {
	if l.svcCtx.RiskBlacklist == nil {
		return nil
	}
	entry, err := l.svcCtx.RiskBlacklist.FindActiveByTarget(l.ctx, "user", userID)
	if err != nil {
		l.Logger.Errorf("query login blacklist failed, user_id=%d: %v", userID, err)
		return nil
	}
	if entry == nil {
		return nil
	}
	return l.svcCtx.RiskBlacklist.CreateHitRecord(l.ctx, &repository.BlacklistHitRecord{
		BlacklistID: entry.ID,
		TargetType:  "user",
		TargetID:    userID,
		Scene:       "login",
		RiskLevel:   3,
		HitReason:   entry.Reason,
		RequestID:   riskRequestID(l.ctx),
	})
}

// riskRequestID 从 gRPC 入站元数据提取请求链路 ID，并限制在运营表字段允许的长度内。
// 现有协议未强制携带该字段，缺失时返回空字符串以兼容历史调用。
func riskRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"x-request-id", "request-id", "trace-id"} {
		for _, value := range md.Get(key) {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if len(value) > 64 {
				return value[:64]
			}
			return value
		}
	}
	return ""
}

// findOrCreateUser 按手机号查询用户；不存在时按短信登录规则自动注册。
func (l *LoginBySMSLogic) findOrCreateUser(phone string) (*model.User, bool, error) {
	user, err := l.svcCtx.Users.FindByPhone(l.ctx, phone)
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		user = &model.User{
			Phone:          phone,
			Nickname:       MaskPhone(phone),
			RegisterSource: model.RegisterSourcePhone,
			Status:         model.UserStatusNormal,
		}
		if err := l.svcCtx.Users.Create(l.ctx, user); err != nil {
			return nil, false, err
		}
		return user, true, nil
	case err != nil:
		return nil, false, err
	default:
		return user, false, nil
	}
}

// toUserInfo 将内部用户模型转换为 proto 响应结构，并避免返回明文手机号。
func toUserInfo(user *model.User) *userproto.UserInfo {
	return &userproto.UserInfo{
		UserId:         user.ID,
		Phone:          MaskPhone(user.Phone),
		Nickname:       user.Nickname,
		AvatarUrl:      user.AvatarURL,
		RealNameStatus: realNameStatus(user),
	}
}
