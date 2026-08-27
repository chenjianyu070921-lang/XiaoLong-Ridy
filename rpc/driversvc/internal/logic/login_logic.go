package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/common/jwtx"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

// 登录相关错误，供上层（grpc 状态码映射）识别。
var (
	ErrLoginAccountNotFound = errors.New("账号不存在")
	ErrLoginPasswordWrong   = errors.New("密码错误")
	ErrLoginAccountBlocked  = errors.New("账号已被冻结或注销")
)

// loginTokenTTL 司机登录令牌有效期。
const loginTokenTTL = 2 * time.Hour

// LoginLogic 司机登录逻辑。
type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewLoginLogic 构造司机登录逻辑处理器。
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 校验手机号、账号状态与密码，成功则签发 JWT 并返回司机简要信息。
func (l *LoginLogic) Login(in *proto.LoginRequest) (*proto.LoginResponse, error) {
	// 按手机号查询司机（不存在时统一返回账号不存在）。
	d, err := l.svcCtx.DriverRepository.GetByPhone(l.ctx, in.GetPhone())
	if err != nil {
		return nil, ErrLoginAccountNotFound
	}
	// 冻结/注销账号拒绝登录。
	if d.Status == int8(proto.DriverStatus_DRIVER_STATUS_FROZEN) ||
		d.Status == int8(proto.DriverStatus_DRIVER_STATUS_CANCELLED) {
		return nil, ErrLoginAccountBlocked
	}
	// 校验 bcrypt 密码哈希。
	if bcrypt.CompareHashAndPassword([]byte(d.PasswordHash), []byte(in.GetPassword())) != nil {
		return nil, ErrLoginPasswordWrong
	}

	// 签发司机 JWT，密钥取自配置，需与 api/driver 端保持一致。
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     uint64(d.Id),
		AccountType:   "driver",
		AccountStatus: int(d.Status),
		Phone:         d.Phone,
		Role:          "driver",
		Issuer:        "driversvc",
		TTL:           loginTokenTTL,
	}, l.svcCtx.Config.SigningKey)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		Token:    token,
		ExpireIn: int64(loginTokenTTL.Seconds()),
		Driver: &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			RealName:        d.RealName,
			IdCardNo:        d.IdCardNo,
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			OnlineStatus:    int32(d.OnlineStatus),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}

// LoginBySMS signs a driver token after the API layer has verified the SMS code.
func (l *LoginLogic) LoginBySMS(in *proto.LoginBySMSRequest) (*proto.LoginResponse, error) {
	d, err := l.svcCtx.DriverRepository.GetByPhone(l.ctx, in.GetPhone())
	if err != nil {
		return nil, ErrLoginAccountNotFound
	}
	if d.Status == int8(proto.DriverStatus_DRIVER_STATUS_FROZEN) ||
		d.Status == int8(proto.DriverStatus_DRIVER_STATUS_CANCELLED) {
		return nil, ErrLoginAccountBlocked
	}

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     uint64(d.Id),
		AccountType:   "driver",
		AccountStatus: int(d.Status),
		Phone:         d.Phone,
		Role:          "driver",
		Issuer:        "driversvc",
		TTL:           loginTokenTTL,
	}, l.svcCtx.Config.SigningKey)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		Token:    token,
		ExpireIn: int64(loginTokenTTL.Seconds()),
		Driver: &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			RealName:        d.RealName,
			IdCardNo:        d.IdCardNo,
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			OnlineStatus:    int32(d.OnlineStatus),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}
