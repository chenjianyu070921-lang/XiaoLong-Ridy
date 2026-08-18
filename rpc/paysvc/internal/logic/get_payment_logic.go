package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentLogic {
	return &GetPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 支付查询：按支付单号（优先）或订单ID查询支付状态。
func (l *GetPaymentLogic) GetPayment(in *proto.GetPaymentRequest) (*proto.GetPaymentResponse, error) {
	repo := repository.NewPaymentRepo(l.svcCtx.DB)

	var (
		p   *model.Payment
		err error
	)
	if in.PaymentNo != "" {
		p, err = repo.FindByPaymentNo(l.ctx, in.PaymentNo)
	} else {
		p, err = repo.FindByOrderId(l.ctx, uint64(in.OrderId))
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	return &proto.GetPaymentResponse{
		PaymentId:         int64(p.Id),
		PaymentNo:         p.PaymentNo,
		OrderId:           int64(p.OrderId),
		AmountCents:       priceutil.YuanToCents(p.Amount),
		Channel:           p.Channel,
		Status:            int32(p.Status),
		TransactionId:     p.TransactionId,
		RefundAmountCents: priceutil.YuanToCents(p.RefundAmount),
	}, nil
}
