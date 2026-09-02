package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePaymentLogic {
	return &CreatePaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// channelName 将 proto 渠道枚举转为渠道标识字符串。
func channelName(c proto.PayChannel) string {
	switch c {
	case proto.PayChannel_PAY_CHANNEL_ALIPAY:
		return channel.Alipay
	case proto.PayChannel_PAY_CHANNEL_BALANCE:
		return channel.Balance
	default:
		return channel.Wechat
	}
}

// genPaymentNo 生成平台支付单号：PAY + 时间戳 + 微秒。
func genPaymentNo() string {
	return fmt.Sprintf("PAY%s%06d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1000)
}

// 支付预下单：创建支付单并调第三方下单（本期为 mock）。
func (l *CreatePaymentLogic) CreatePayment(in *proto.CreatePaymentRequest) (*proto.CreatePaymentResponse, error) {
	// 1. 生成支付单号
	paymentNo := genPaymentNo()
	chName := channelName(in.Channel)

	// 2. 创建支付单（待支付）
	// 余额支付在当前系统内无需第三方回调，扣款请求成功即可完成支付。
	initialStatus := int8(model.PaymentStatusPending)
	if chName == channel.Balance {
		initialStatus = model.PaymentStatusPaid
	}
	payment := &model.Payment{
		PaymentNo: paymentNo,
		OrderId:   uint64(in.OrderId),
		UserId:    uint64(in.UserId),
		Amount:    priceutil.CentsToYuan(in.AmountCents),
		Channel:   chName,
		Status:    initialStatus,
	}
	if initialStatus == model.PaymentStatusPaid {
		now := time.Now()
		payment.PaidAt = &now
	}
	repo := repository.NewPaymentRepo(l.svcCtx.DB)
	if err := repo.Create(l.ctx, payment); err != nil {
		return nil, err
	}

	// 3. 调渠道下单（本期 mock）
	ch := channel.NewMockChannel(chName)
	result, err := ch.CreateOrder(l.ctx, paymentNo, in.AmountCents)
	if err != nil {
		return nil, err
	}

	// 4. 回填第三方流水号
	payment.TransactionId = result.TransactionId
	if err := repo.Update(l.ctx, payment); err != nil {
		return nil, err
	}
	if initialStatus == model.PaymentStatusPaid && l.svcCtx.OrderClient != nil {
		// 跨服务不能共享同一数据库事务，因此采用有限重试补偿确认，避免支付成功但订单仍待支付。
		confirmReq := &orderproto.ConfirmPaidRequest{OrderId: in.OrderId, PaymentNo: paymentNo, AmountCents: in.AmountCents, PaidAt: time.Now().Unix()}
		var confirmErr error
		for attempt := 0; attempt < 3; attempt++ {
			if _, confirmErr = l.svcCtx.OrderClient.ConfirmPaid(l.ctx, confirmReq); confirmErr == nil {
				break
			}
			time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
		}
		if confirmErr != nil {
			l.Logger.Errorf("余额支付已成功但订单确认失败，将由后续支付查询补偿: order_id=%d payment_no=%s err=%v", in.OrderId, paymentNo, confirmErr)
		}
	}

	return &proto.CreatePaymentResponse{
		PaymentId:     int64(payment.Id),
		PaymentNo:     paymentNo,
		TransactionId: result.TransactionId,
		PayParams:     result.PayParams,
		Status:        int32(initialStatus),
	}, nil
}
