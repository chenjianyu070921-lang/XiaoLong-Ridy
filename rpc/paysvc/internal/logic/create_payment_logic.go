package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/errorx"
	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 错误标识：渠道下单成功但回写 transaction_id 失败，可对账补单。
var errCreatePaymentFinalize = errors.New("create payment finalize failed")

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

// CreatePayment 支付预下单：创建支付单并调第三方下单（本期为 mock，开发态 mock，生产态按配置启用 AlipayChannel）。
//
// 事务边界：
//   - DB 写入（Create payment + 回填 transaction_id）包在一个事务里，
//     防止"第三方已下单但 transaction_id 落库失败"造成对账黑洞。
//   - 渠道下单放在事务外：第三方 RPC 是长耗时操作，包在事务里会持续占连接。
func (l *CreatePaymentLogic) CreatePayment(in *proto.CreatePaymentRequest) (*proto.CreatePaymentResponse, error) {
	// 入参校验：金额必须为正。
	if in.AmountCents <= 0 {
		return nil, errorx.NewErr(errorx.PARAM)
	}

	// 1. 生成支付单号
	paymentNo := genPaymentNo()
	chName := channelName(in.Channel)

	// 2. 调渠道下单（放在事务外，避免持锁做 RPC）。
	ch := l.svcCtx.GetChannel(chName)
	channelResult, err := ch.CreateOrder(l.ctx, paymentNo, in.AmountCents)
	if err != nil {
		return nil, fmt.Errorf("channel create order failed: %w", err)
	}

	// 3. DB 写入：用事务包"创建支付单 + 回填 transaction_id"。
	//    闭包返回值用于事务外构造响应。
	var paymentID int64
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		payment := &model.Payment{
			PaymentNo:     paymentNo,
			OrderId:       uint64(in.OrderId),
			UserId:        uint64(in.UserId),
			Amount:        priceutil.CentsToYuan(in.AmountCents),
			Channel:       chName,
			Status:        model.PaymentStatusPending, // 待支付
			TransactionId: channelResult.TransactionId,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}
		paymentID = int64(payment.Id)
		return nil
	})
	if err != nil {
		// 这里渠道已经下单成功但 DB 没写成功，需要后续对账补单；先回传 500 让上游感知。
		return nil, fmt.Errorf("%w: %v", errCreatePaymentFinalize, err)
	}

	return &proto.CreatePaymentResponse{
		PaymentId:     paymentID,
		PaymentNo:     paymentNo,
		TransactionId: channelResult.TransactionId,
		PayParams:     channelResult.PayParams,
		Status:        int32(model.PaymentStatusPending),
	}, nil
}
