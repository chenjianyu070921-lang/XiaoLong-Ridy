package logic

import (
	"context"
		"strings"

	"XiaoLong-Ridy/common/cryptox"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ResetWithdrawPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetWithdrawPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetWithdrawPasswordLogic {
	return &ResetWithdrawPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ResetWithdrawPassword 遗忘提现密码找回：仅凭「手机号 + 身份证号 + 真实姓名」三重校验通过后，
// 重置为新的 6 位数字提现密码（bcrypt 存储，明文仅本响应返回一次，调用方只用于短信下发）。
func (l *ResetWithdrawPasswordLogic) ResetWithdrawPassword(in *__proto.ResetWithdrawPasswordRequest) (*__proto.ResetWithdrawPasswordResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数不能为空")
	}
	phone := strings.TrimSpace(in.GetPhone())
	idCardNo := strings.TrimSpace(in.GetIdCardNo())
	realName := strings.TrimSpace(in.GetRealName())
	if phone == "" || idCardNo == "" || realName == "" {
		return nil, status.Error(codes.InvalidArgument, "手机号、身份证号、真实姓名均不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil || l.svcCtx.DriverBankCardRepository == nil {
		return nil, status.Error(codes.InvalidArgument, "driver repository not ready")
	}

	driver, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "司机不存在")
	}
	if driver.Phone != phone || driver.IdCardNo != idCardNo || driver.RealName != realName {
		return nil, status.Error(codes.InvalidArgument, "手机号、身份证号或真实姓名与实名信息不符")
	}

	plainPassword, err := randomDigits(withdrawPasswordLen)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "提现密码生成失败")
	}
	hash, err := cryptox.BcryptHash(plainPassword)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "提现密码加密失败")
	}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.GetDriverId()), map[string]interface{}{
		"withdraw_password_hash": hash,
	}); err != nil {
		return nil, err
	}
	// 冗余同步到全部已绑银行卡，保证按卡读取时密码一致。
	if err := l.svcCtx.DriverBankCardRepository.UpdateWithdrawPasswordHash(l.ctx, uint64(in.GetDriverId()), hash); err != nil {
		return nil, err
	}
	return &__proto.ResetWithdrawPasswordResponse{WithdrawPassword: plainPassword}, nil
}
