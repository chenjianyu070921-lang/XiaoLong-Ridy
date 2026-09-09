package logic

import (
	"context"
	"errors"
	"log"
	"strings"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// 合规红线：绑卡全程不采集银行卡取款密码；平台提现密码仅用于提现本人确认。
// 绑卡短信验证码以 "bank:" 前缀区分于登录验证码，避免同手机号场景下互相覆盖。

// BankCardLogic 封装司机银行卡与平台提现密码逻辑。
type BankCardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankCardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BankCardLogic {
	return &BankCardLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendBankCardSmsCode 向银行预留手机号发送绑卡验证码（联调阶段打印到日志）。
func (l *BankCardLogic) SendBankCardSmsCode(req *types.SendBankCardSmsCodeRequest) error {
	if req == nil {
		return ErrInvalidParam
	}
	phone := strings.TrimSpace(req.ReservedPhone)
	if !validPhone(phone) {
		return errors.New("银行预留手机号格式不合法")
	}
	if l.svcCtx == nil || l.svcCtx.CodeCache == nil {
		return ErrCodeSendFailed
	}
	code, err := randomNumericCode(6)
	if err != nil {
		return ErrCodeSendFailed
	}
	l.svcCtx.CodeCache.Set(bankCodeKey(phone), code)
	logSMS(phone, code)
	return nil
}

// BindBankCard 绑卡：先校验银行预留手机号短信验证码，再委托 driversvc 完成
// 实名一致、数量上限、卡号加密入库与提现密码生成；新提现密码仅短信下发，不返回明文。
func (l *BankCardLogic) BindBankCard(driverID int64, req *types.BindBankCardRequest) (*types.BindBankCardResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	reservedPhone := strings.TrimSpace(req.ReservedPhone)
	if l.svcCtx == nil || l.svcCtx.CodeCache == nil {
		return nil, ErrCodeInvalid
	}
	if !l.svcCtx.CodeCache.Verify(bankCodeKey(reservedPhone), strings.TrimSpace(req.SmsCode)) {
		return nil, ErrCodeInvalid
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.BindBankCard(l.ctx, &driversproto.BindBankCardRequest{
		DriverId:     driverID,
		BankName:     strings.TrimSpace(req.BankName),
		CardNo:       strings.TrimSpace(req.CardNo),
		HolderName:   strings.TrimSpace(req.HolderName),
		HolderIdCard: strings.TrimSpace(req.HolderIdCard),
		ReservedPhone: reservedPhone,
	})
	if err != nil {
		return nil, err
	}
	// 首次绑卡自动生成的平台提现密码：仅短信下发一次，不进入响应体。
	if pwd := resp.GetWithdrawPassword(); pwd != "" {
		l.notifyWithdrawPassword(driverID, pwd)
	}
	return &types.BindBankCardResponse{
		ID:           resp.GetId(),
		BankName:     resp.GetBankName(),
		MaskedCardNo: resp.GetMaskedCardNo(),
	}, nil
}

// ListBankCards 列出司机已绑银行卡（脱敏视图）。
func (l *BankCardLogic) ListBankCards(driverID int64) (*types.ListBankCardsResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListBankCards(l.ctx, &driversproto.ListBankCardsRequest{DriverId: driverID})
	if err != nil {
		return nil, err
	}
	result := &types.ListBankCardsResponse{}
	for _, card := range resp.GetCards() {
		result.Cards = append(result.Cards, types.BankCardInfo{
			ID:            card.GetId(),
			BankName:      card.GetBankName(),
			MaskedCardNo:  card.GetMaskedCardNo(),
			HolderName:    card.GetHolderName(),
			ReservedPhone: card.GetReservedPhone(),
			CreatedAt:     card.GetCreatedAt(),
		})
	}
	return result, nil
}

// DeleteBankCard 删除司机银行卡。
func (l *BankCardLogic) DeleteBankCard(driverID int64, req *types.DeleteBankCardRequest) error {
	if driverID <= 0 || req == nil || req.ID <= 0 {
		return ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return err
	}
	_, err = client.DeleteBankCard(l.ctx, &driversproto.DeleteBankCardRequest{
		DriverId: driverID,
		Id:       req.ID,
	})
	return err
}

// ResetWithdrawPassword 遗忘提现密码找回：driversvc 内做「手机号+身份证+真实姓名」三重校验，
// 重置成功后新密码仅短信下发，不返回明文。
func (l *BankCardLogic) ResetWithdrawPassword(driverID int64, req *types.ResetWithdrawPasswordRequest) (*types.ResetWithdrawPasswordResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.ResetWithdrawPassword(l.ctx, &driversproto.ResetWithdrawPasswordRequest{
		DriverId:  driverID,
		Phone:     strings.TrimSpace(req.Phone),
		IdCardNo:  strings.TrimSpace(req.IdCardNo),
		RealName:  strings.TrimSpace(req.RealName),
	})
	if err != nil {
		return nil, err
	}
	if pwd := resp.GetWithdrawPassword(); pwd != "" {
		l.notifyWithdrawPassword(driverID, pwd)
	}
	return &types.ResetWithdrawPasswordResponse{Message: "新提现密码已发送至注册手机号，请注意查收"}, nil
}

// notifyWithdrawPassword 将平台提现密码发送至司机注册手机号（联调阶段打印到日志）。
func (l *BankCardLogic) notifyWithdrawPassword(driverID int64, password string) {
	phone := ""
	if client, err := l.driverClient(); err == nil {
		if resp, err := client.GetDriver(l.ctx, &driversproto.GetDriverRequest{Id: driverID}); err == nil && resp.GetDriver() != nil {
			phone = resp.GetDriver().GetPhone()
		}
	}
	if phone == "" {
		log.Printf("[driver-bank] 平台提现密码生成 driver_id=%d（手机号未知，未下发）", driverID)
		return
	}
	logSMS(phone, "平台提现密码:"+password)
}

func (l *BankCardLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}

// bankCodeKey 绑卡验证码缓存 key，与登录验证码隔离。
func bankCodeKey(phone string) string {
	return "bank:" + strings.TrimSpace(phone)
}
