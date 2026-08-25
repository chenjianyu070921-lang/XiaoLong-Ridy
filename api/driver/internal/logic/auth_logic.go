package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/jwtx"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrDriverAuthFailed 表示账号或密码错误、账号不可用等登录失败。
	ErrDriverAuthFailed = errors.New("账号或密码错误")
	// ErrDriverFrozen 表示司机账号未审核通过、已冻结或注销，不允许登录。
	ErrDriverFrozen = errors.New("账号未审核通过或已被冻结/注销")
	// ErrCodeInvalid 表示验证码错误。
	ErrCodeInvalid = errors.New("验证码错误")
	// ErrCodeSendFailed 表示验证码生成失败。
	ErrCodeSendFailed = errors.New("验证码发送失败")
)

// AuthLogic 封装司机登录、发码等认证逻辑。
type AuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAuthLogic 构造司机认证逻辑处理器。
func NewAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendSMSCode 生成验证码并存入本地缓存（联调阶段顶替短信通道）。
func (l *AuthLogic) SendSMSCode(req *types.SendSMSCodeRequest) (*types.SendSMSCodeResponse, error) {
	// 校验手机号格式。
	if !validPhone(strings.TrimSpace(req.Phone)) {
		return nil, errors.New("手机号格式不合法")
	}
	// 生成 6 位随机数字验证码。
	code, err := randomNumericCode(6)
	if err != nil {
		return nil, ErrCodeSendFailed
	}
	// 存入本地验证码缓存。
	l.svcCtx.CodeCache.Set(req.Phone, code)
	// 联调阶段将验证码打印到日志，方便用 curl/Postman 获取（顶替真实短信）。
	logSMS(req.Phone, code)
	return &types.SendSMSCodeResponse{
		Success:  true,
		ExpireIn: int(l.svcCtx.CodeCache.TTL().Seconds()),
	}, nil
}

// LoginByPassword 手机号 + 密码登录，委托 driversvc 校验账号状态与密码并签发 JWT。
func (l *AuthLogic) LoginByPassword(req *types.LoginByPasswordRequest) (*types.LoginResponse, error) {
	if req == nil || !validPhone(strings.TrimSpace(req.Phone)) {
		return nil, ErrDriverAuthFailed
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Login(l.ctx, &driversproto.LoginRequest{
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
	})
	if err != nil {
		return nil, normalizeLoginError(err)
	}
	return toLoginResponse(resp), nil
}

// LoginBySMS 手机号 + 验证码登录；验证码由 API 校验，账号状态与 JWT 签发委托 driversvc。
func (l *AuthLogic) LoginBySMS(req *types.LoginBySMSRequest) (*types.LoginResponse, error) {
	if req == nil || !validPhone(strings.TrimSpace(req.Phone)) {
		return nil, ErrDriverAuthFailed
	}
	if !l.svcCtx.CodeCache.Verify(req.Phone, strings.TrimSpace(req.Code)) {
		return nil, ErrCodeInvalid
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.LoginBySMS(l.ctx, &driversproto.LoginBySMSRequest{Phone: strings.TrimSpace(req.Phone)})
	if err != nil {
		return nil, normalizeLoginError(err)
	}
	return toLoginResponse(resp), nil
}

func toLoginResponse(resp *driversproto.LoginResponse) *types.LoginResponse {
	d := resp.GetDriver()
	return &types.LoginResponse{
		Token:    resp.GetToken(),
		ExpireIn: resp.GetExpireIn(),
		Driver: types.DriverBrief{
			ID:     d.GetId(),
			Phone:  jwtx.MaskPhone(d.GetPhone()),
			Status: d.GetStatus().String(),
		},
	}
}

func normalizeLoginError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		if isDriverForbiddenLoginMessage(err.Error()) {
			return ErrDriverFrozen
		}
		return ErrDriverAuthFailed
	}
	switch st.Code() {
	case codes.PermissionDenied:
		return ErrDriverFrozen
	case codes.Unavailable:
		return err
	default:
		if isDriverForbiddenLoginMessage(st.Message()) {
			return ErrDriverFrozen
		}
		return ErrDriverAuthFailed
	}
}

func isDriverForbiddenLoginMessage(message string) bool {
	return strings.Contains(message, "未审核") ||
		strings.Contains(message, "冻结") ||
		strings.Contains(message, "注销") ||
		strings.Contains(message, "blocked") ||
		strings.Contains(message, "unavailable")
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *AuthLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

// randomNumericCode 生成指定长度的随机数字验证码（字符串形式）。
func randomNumericCode(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const digits = "0123456789"
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = digits[int(b)%len(digits)]
	}
	return string(out), nil
}

// logSMS 在联调阶段将验证码输出到日志，顶替真实短信通道。
func logSMS(phone, code string) {
	log.Printf("[driver-auth] 本地短信验证码 phone=%s code=%s", phone, code)
}
