package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	"XiaoLong-Ridy/common/jwtx"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"google.golang.org/grpc/status"
)

var (
	// ErrDriverAuthFailed 表示账号或密码错误、账号不可用等登录失败。
	ErrDriverAuthFailed = errors.New("账号或密码错误")
	// ErrDriverFrozen 表示司机账号已被冻结或注销，不允许登录。
	ErrDriverFrozen = errors.New("账号已被冻结或注销")
	// ErrCodeInvalid 表示验证码错误。
	ErrCodeInvalid = errors.New("验证码错误")
	// ErrCodeSendFailed 表示验证码生成失败。
	ErrCodeSendFailed = errors.New("验证码发送失败")
)

// accessTokenTTL 司机登录令牌有效期。
const accessTokenTTL = 2 * time.Hour

// AuthLogic 封装司机登录、发码等认证逻辑。
type AuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAuthLogic 构造司机认证逻辑处理器。
func NewAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendSMSCode 处理「发送登录短信验证码」请求：校验手机号格式，生成 6 位随机验证码存入本地缓存，
// 联调阶段将验证码打印到日志顶替真实短信通道，返回验证码有效期。
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

// LoginByPassword 处理「手机号 + 密码」登录：调用 driversvc.Login 完成账号状态与密码校验，
// 下游返回的错误统一映射为登录失败（不泄露细节），成功后透传 JWT 与脱敏司机信息。
func (l *AuthLogic) LoginByPassword(req *types.LoginByPasswordRequest) (*types.LoginResponse, error) {
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.Login(l.ctx, &driversproto.LoginRequest{
		Phone:    strings.TrimSpace(req.Phone),
		Password: req.Password,
	})
	if err != nil {
		// driversvc 返回的账号/密码错误统一映射为登录失败，避免泄露细节。
		if _, ok := status.FromError(err); ok {
			return nil, ErrDriverAuthFailed
		}
		return nil, err
	}
	d := resp.GetDriver()
	return &types.LoginResponse{
		Token:    resp.GetToken(),
		ExpireIn: resp.GetExpireIn(),
		Driver: types.DriverBrief{
			ID:     d.GetId(),
			Phone:  jwtx.MaskPhone(d.GetPhone()),
			Status: d.GetStatus().String(),
		},
	}, nil
}

// LoginBySMS 处理「手机号 + 短信验证码」登录：先校验验证码（无论司机是否存在都校验，避免泄露账号存在性），
// 再按手机号加载司机并校验状态，全部通过后签发 JWT。
func (l *AuthLogic) LoginBySMS(req *types.LoginBySMSRequest) (*types.LoginResponse, error) {
	// 先校验验证码（无论司机是否存在都校验，避免泄露账号存在性）。
	if !l.svcCtx.CodeCache.Verify(req.Phone, strings.TrimSpace(req.Code)) {
		return nil, ErrCodeInvalid
	}
	driver, err := l.loadDriverByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	return l.issueToken(driver)
}

// loadDriverByPhone 按手机号加载司机实体，并做基础校验：手机号格式、存在性、是否冻结/注销。
// 任一不通过均返回登录失败，避免向下游或调用方泄露具体原因。
func (l *AuthLogic) loadDriverByPhone(phone string) (*driversproto.Driver, error) {
	if !validPhone(strings.TrimSpace(phone)) {
		return nil, ErrDriverAuthFailed
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetDriverByPhone(l.ctx, &driversproto.GetDriverByPhoneRequest{Phone: strings.TrimSpace(phone)})
	if err != nil {
		// 未找到司机或下游异常，统一返回登录失败，避免泄露细节。
		return nil, ErrDriverAuthFailed
	}
	d := resp.GetDriver()
	// 冻结/注销账号拒绝登录。
	if d.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_FROZEN ||
		d.GetStatus() == driversproto.DriverStatus_DRIVER_STATUS_CANCELLED {
		return nil, ErrDriverFrozen
	}
	return d, nil
}

// issueToken 为给定司机实体签发 JWT（HMAC-SHA256，有效期 2 小时），并组装统一登录响应。
func (l *AuthLogic) issueToken(d *driversproto.Driver) (*types.LoginResponse, error) {
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     uint64(d.GetId()),
		AccountType:   "driver",
		AccountStatus: int(d.GetStatus()),
		Phone:         d.GetPhone(),
		Role:          "driver",
		Issuer:        "driver-api",
		TTL:           accessTokenTTL,
	}, l.svcCtx.SigningKey)
	if err != nil {
		return nil, err
	}
	return &types.LoginResponse{
		Token:    token,
		ExpireIn: int64(accessTokenTTL.Seconds()),
		Driver: types.DriverBrief{
			ID:     d.GetId(),
			Phone:  jwtx.MaskPhone(d.GetPhone()),
			Status: d.GetStatus().String(),
		},
	}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端；未配置时返回 ErrDriverClientNotConfigured。
func (l *AuthLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

// randomNumericCode 生成长度为 n 的随机数字验证码（每位 0-9 均匀随机），以字符串形式返回。
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

// logSMS 在联调阶段将验证码打印到服务日志，顶替真实短信通道，方便用 curl/Postman 直接获取。
func logSMS(phone, code string) {
	log.Printf("[driver-auth] 本地短信验证码 phone=%s code=%s", phone, code)
}
