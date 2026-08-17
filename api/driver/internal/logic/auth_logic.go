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
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	"XiaoLong-Ridy/common/jwtx"

	"golang.org/x/crypto/bcrypt"
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

// LoginByPassword 手机号 + 密码登录，校验通过后签发 JWT。
func (l *AuthLogic) LoginByPassword(req *types.LoginByPasswordRequest) (*types.LoginResponse, error) {
	driver, err := l.loadDriverByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	// 比对 bcrypt 密码哈希。
	if bcrypt.CompareHashAndPassword([]byte(driver.GetPasswordHash()), []byte(req.Password)) != nil {
		return nil, ErrDriverAuthFailed
	}
	return l.issueToken(driver)
}

// LoginBySMS 手机号 + 验证码登录，校验通过后签发 JWT。
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

// loadDriverByPhone 通过手机号查询司机，并完成账号状态基础校验。
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

// issueToken 为司机签发 JWT 并构造登录响应。
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
