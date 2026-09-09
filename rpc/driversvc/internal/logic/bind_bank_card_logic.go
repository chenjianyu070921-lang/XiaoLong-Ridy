package logic

import (
	"context"
	"crypto/rand"
		"strings"

	"XiaoLong-Ridy/common/cryptox"
	"XiaoLong-Ridy/rpc/driversvc/internal/crypto"
	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 合规红线：绑卡全程不采集银行卡取款密码；平台提现密码为独立密码，仅用于平台提现本人确认。
const (
	// maxBankCardsPerDriver 单个司机最多绑定的银行卡数量。
	maxBankCardsPerDriver = 5
	// withdrawPasswordLen 平台提现密码长度（6 位数字）。
	withdrawPasswordLen = 6
)

type BindBankCardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBindBankCardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BindBankCardLogic {
	return &BindBankCardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BindBankCard 绑卡：校验实名一致、数量上限与重复卡，卡号 AES 加密入库；
// 司机尚无平台提现密码时自动生成 6 位数字密码（bcrypt 存储，明文仅本响应返回一次，调用方只用于短信下发）。
func (l *BindBankCardLogic) BindBankCard(in *__proto.BindBankCardRequest) (*__proto.BindBankCardResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverBankCardRepository == nil || l.svcCtx.DriverRepository == nil {
		return nil, status.Error(codes.InvalidArgument, "driver bank card repository not ready")
	}
	bankName := strings.TrimSpace(in.GetBankName())
	cardNo := strings.TrimSpace(in.GetCardNo())
	holderName := strings.TrimSpace(in.GetHolderName())
	holderIdCard := strings.TrimSpace(in.GetHolderIdCard())
	reservedPhone := strings.TrimSpace(in.GetReservedPhone())
	if bankName == "" || cardNo == "" || holderName == "" || holderIdCard == "" || reservedPhone == "" {
		return nil, status.Error(codes.InvalidArgument, "卡号、开户行、姓名、身份证、银行预留手机号均不能为空")
	}
	if !validBankCardNo(cardNo) {
		return nil, status.Error(codes.InvalidArgument, "银行卡号格式不正确")
	}

	driverID := uint64(in.GetDriverId())

	// 实名一致校验：持卡人姓名与身份证必须与司机实名完全一致。
	driver, err := l.svcCtx.DriverRepository.GetByID(l.ctx, driverID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "司机不存在")
	}
	if driver.RealName != holderName || driver.IdCardNo != holderIdCard {
		return nil, status.Error(codes.InvalidArgument, "持卡人信息必须与司机实名信息一致")
	}

	// 数量上限：最多绑定 5 张卡。
	count, err := l.svcCtx.DriverBankCardRepository.CountByDriver(l.ctx, driverID)
	if err != nil {
		return nil, err
	}
	if count >= maxBankCardsPerDriver {
		return nil, status.Error(codes.InvalidArgument, "最多绑定5张银行卡，如需绑定新卡请先删除已有卡片")
	}

	// 重复卡校验（同司机同卡号不可重复绑定）。
	cardHash := crypto.CardNoHash(cardNo)
	exists, err := l.svcCtx.DriverBankCardRepository.ListByDriver(l.ctx, driverID)
	if err != nil {
		return nil, err
	}
	for _, card := range exists {
		if card.CardNoHash == cardHash {
			return nil, status.Error(codes.InvalidArgument, "该银行卡已绑定，请勿重复添加")
		}
	}

	// 平台提现密码：司机首次绑卡时自动生成。
	var plainPassword string
	if driver.WithdrawPasswordHash == "" {
		plainPassword, err = randomDigits(withdrawPasswordLen)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "提现密码生成失败")
		}
		hash, err := cryptox.BcryptHash(plainPassword)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "提现密码加密失败")
		}
		if err := l.svcCtx.DriverRepository.Update(l.ctx, driverID, map[string]interface{}{
			"withdraw_password_hash": hash,
		}); err != nil {
			return nil, err
		}
	}

	// 卡号 AES 加密后入库。
	encrypted, err := crypto.EncryptCardNo(cardNo, l.svcCtx.Config.CardKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "卡号加密失败")
	}
	card := &model.DriverBankCard{
		DriverId:             driverID,
		BankName:             bankName,
		CardNo:               encrypted,
		CardNoHash:           cardHash,
		HolderName:           holderName,
		HolderIdCard:         holderIdCard,
		ReservedPhone:        reservedPhone,
		WithdrawPasswordHash: driver.WithdrawPasswordHash,
		Status:               1,
	}
	if err := l.svcCtx.DriverBankCardRepository.Create(l.ctx, card); err != nil {
		return nil, err
	}

	return &__proto.BindBankCardResponse{
		Id:               int64(card.Id),
		BankName:         bankName,
		MaskedCardNo:     crypto.MaskCardNo(cardNo),
		WithdrawPassword: plainPassword,
	}, nil
}

// validBankCardNo 校验银行卡号：8-19 位纯数字（借记卡/信用卡通用规则）。
func validBankCardNo(cardNo string) bool {
	if len(cardNo) < 8 || len(cardNo) > 19 {
		return false
	}
	for _, r := range cardNo {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// randomDigits 生成指定长度的随机数字串。
func randomDigits(n int) (string, error) {
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
