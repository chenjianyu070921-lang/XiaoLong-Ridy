package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/errorx"
	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
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

// CreatePayment 支付预下单：先落库 pending 支付单，再调第三方渠道下单并回填 transaction_id。
//
// 事务边界（解决 M5-6 对账黑洞）：
//   - 阶段 1：DB 事务内先写入 status=pending 的支付单（无 transaction_id），保证后续可对账；
//   - 阶段 2：在事务外调用第三方渠道下单（避免长事务持锁）；
//   - 阶段 3：渠道成功则 DB 事务回填 transaction_id；渠道失败则把支付单标记为 failed。
//   - 生产环境真实渠道不能放入事务，因此靠"先落库 + 状态机 + 对账任务"保证最终一致。
func (l *CreatePaymentLogic) CreatePayment(in *proto.CreatePaymentRequest) (*proto.CreatePaymentResponse, error) {
	// 入参校验：金额必须为正。
	if in.AmountCents <= 0 {
		return nil, errorx.NewErr(errorx.PARAM)
	}

	// 1. 生成支付单号
	paymentNo := genPaymentNo()
	chName := channelName(in.Channel)

	// 2. 先落库 pending 支付单（无 transaction_id），确保后续可对账。
	payment := &model.Payment{
		PaymentNo:   paymentNo,
		OrderId:     uint64(in.OrderId),
		UserId:      uint64(in.UserId),
		AmountCents: in.AmountCents,
		Channel:     chName,
		Status:      model.PaymentStatusPending,
	}
	repo := repository.NewPaymentRepo(l.svcCtx.DB)
	if err := repo.Create(l.ctx, payment); err != nil {
		return nil, fmt.Errorf("create payment failed: %w", err)
	}

	// 3. 在事务外调用第三方渠道下单（mock / alipay）。
	ch := l.svcCtx.GetChannel(chName)
	channelResult, err := ch.CreateOrder(l.ctx, paymentNo, in.AmountCents)
	if err != nil {
		// 渠道下单失败：把支付单标记为失败，避免这笔单长期处于 pending 而形成对账干扰。
		if updateErr := repo.UpdateSelective(l.ctx, payment.Id, map[string]interface{}{
			"status": model.PaymentStatusFailed,
		}); updateErr != nil {
			l.Errorf("mark payment failed after channel error: %v", updateErr)
		}
		return nil, fmt.Errorf("channel create order failed: %w", err)
	}

	// 4. 回填第三方交易流水号，严格限定只在 pending 状态下更新，防止并发/重试时误改已支付单。
	if err := repo.UpdateSelective(l.ctx, payment.Id, map[string]interface{}{
		"transaction_id": channelResult.TransactionId,
	}); err != nil {
		// 渠道已下单但回填失败，DB 里已存在 pending 单，可通过对账任务根据 payment_no 补填 transaction_id。
		return nil, fmt.Errorf("%w: %v", errCreatePaymentFinalize, err)
	}

	return &proto.CreatePaymentResponse{
		PaymentId:     int64(payment.Id),
		PaymentNo:     paymentNo,
		TransactionId: channelResult.TransactionId,
		PayParams:     channelResult.PayParams,
		Status:        int32(model.PaymentStatusPending),
	}, nil
}
