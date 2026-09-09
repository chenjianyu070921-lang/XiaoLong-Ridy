package logic

import (
	"context"
	
	"XiaoLong-Ridy/common/cryptox"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VerifyWithdrawPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyWithdrawPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyWithdrawPasswordLogic {
	return &VerifyWithdrawPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// VerifyWithdrawPassword 校验平台提现密码（bcrypt），用于提现操作前的本人确认。
func (l *VerifyWithdrawPasswordLogic) VerifyWithdrawPassword(in *__proto.VerifyWithdrawPasswordRequest) (*__proto.CommonResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数不能为空")
	}
	if in.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "请输入平台提现密码")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, status.Error(codes.InvalidArgument, "driver repository not ready")
	}
	driver, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "司机不存在")
	}
	if driver.WithdrawPasswordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "尚未设置平台提现密码，请先绑定银行卡")
	}
	if !cryptox.BcryptCompare(in.GetPassword(), driver.WithdrawPasswordHash) {
		return nil, status.Error(codes.InvalidArgument, "平台提现密码错误")
	}
	return &__proto.CommonResponse{Message: "ok"}, nil
}
