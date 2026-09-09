package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/crypto"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBankCardsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBankCardsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBankCardsLogic {
	return &ListBankCardsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListBankCards 列出司机已绑银行卡：卡号仅返回前3后3脱敏视图，不返回任何敏感原文。
func (l *ListBankCardsLogic) ListBankCards(in *__proto.ListBankCardsRequest) (*__proto.ListBankCardsResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errors.New("请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverBankCardRepository == nil {
		return nil, errors.New("driver bank card repository not ready")
	}
	cards, err := l.svcCtx.DriverBankCardRepository.ListByDriver(l.ctx, uint64(in.GetDriverId()))
	if err != nil {
		return nil, err
	}
	resp := &__proto.ListBankCardsResponse{}
	for _, card := range cards {
		// 脱敏失败（密钥变更等）时返回空串，绝不泄露原文。
		masked := ""
		if plain, err := crypto.DecryptCardNo(card.CardNo, l.svcCtx.Config.CardKey); err == nil {
			masked = crypto.MaskCardNo(plain)
		}
		resp.Cards = append(resp.Cards, &__proto.BankCardInfo{
			Id:           int64(card.Id),
			BankName:     card.BankName,
			MaskedCardNo: masked,
			HolderName:   card.HolderName,
			ReservedPhone: card.ReservedPhone,
			CreatedAt:    card.CreatedAt.Unix(),
		})
	}
	return resp, nil
}
