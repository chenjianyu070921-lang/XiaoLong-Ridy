package logic

import (
	"context"
		"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/common/cryptox"
	"XiaoLong-Ridy/rpc/driversvc/internal/crypto"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// withdrawStatusPending 提现申请中（司机侧发起后的初始状态）。
const withdrawStatusPending int8 = 1

type CreateWithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawLogic {
	return &CreateWithdrawLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateWithdraw 创建提现申请：必须选择本人已绑银行卡并输入平台提现密码；
// 收款账户取银行卡信息（姓名与实名一致、卡号由密文解密），初始状态为申请中。
// 注：实际打款由 adminsvc 审核后执行，本 RPC 仅负责司机侧发起与记录。
func (l *CreateWithdrawLogic) CreateWithdraw(in *proto.CreateWithdrawRequest) (*proto.CreateWithdrawResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverWithdrawRepository == nil || l.svcCtx.DriverBankCardRepository == nil {
		return nil, status.Error(codes.InvalidArgument, "driver withdraw repository not ready")
	}
	if err := validateWithdrawAmount(in.Amount); err != nil {
		return nil, err
	}
	if in.BankCardId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请选择提现银行卡")
	}
	if strings.TrimSpace(in.WithdrawPassword) == "" {
		return nil, status.Error(codes.InvalidArgument, "请输入平台提现密码")
	}

	driverID := uint64(in.DriverId)

	// 校验银行卡归属。
	card, err := l.svcCtx.DriverBankCardRepository.GetByID(l.ctx, uint64(in.BankCardId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "提现银行卡不存在")
	}
	if card.DriverId != driverID {
		return nil, status.Error(codes.InvalidArgument, "无权使用该银行卡提现")
	}

	// 校验平台提现密码（本人确认）。
	if err := l.verifyPassword(l.ctx, driverID, in.WithdrawPassword); err != nil {
		return nil, err
	}

	// 解密卡号作为收款账户。
	payAccount, err := crypto.DecryptCardNo(card.CardNo, l.svcCtx.Config.CardKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "银行卡信息解密失败")
	}

	now := time.Now()
	withdraw := &model.DriverWithdraw{
		DriverId:   driverID,
		WithdrawNo: generateWithdrawNo(now),
		Amount:     in.Amount,
		PayeeName:  card.HolderName,
		PayAccount: payAccount,
		Status:     withdrawStatusPending,
		AppliedAt:  &now,
	}
	if err := l.svcCtx.DriverWithdrawRepository.Create(l.ctx, withdraw); err != nil {
		return nil, err
	}
	return &proto.CreateWithdrawResponse{
		Id:         int64(withdraw.Id),
		WithdrawNo: withdraw.WithdrawNo,
		Status:     int32(withdraw.Status),
		CreatedAt:  withdraw.CreatedAt.Unix(),
	}, nil
}

// verifyPassword 校验司机平台提现密码（bcrypt）。
func (l *CreateWithdrawLogic) verifyPassword(ctx context.Context, driverID uint64, password string) error {
	if l.svcCtx.DriverRepository == nil {
		return status.Error(codes.InvalidArgument, "driver repository not ready")
	}
	driver, err := l.svcCtx.DriverRepository.GetByID(ctx, driverID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "司机不存在")
	}
	if driver.WithdrawPasswordHash == "" {
		return status.Error(codes.InvalidArgument, "尚未设置平台提现密码，请先绑定银行卡")
	}
	if !cryptox.BcryptCompare(password, driver.WithdrawPasswordHash) {
		return status.Error(codes.InvalidArgument, "平台提现密码错误")
	}
	return nil
}

// generateWithdrawNo 生成提现单号：WD + 时间(秒) + 4位随机数。
func generateWithdrawNo(now time.Time) string {
	return fmt.Sprintf("WD%010d%04d", now.Unix(), now.Nanosecond()%10000)
}
