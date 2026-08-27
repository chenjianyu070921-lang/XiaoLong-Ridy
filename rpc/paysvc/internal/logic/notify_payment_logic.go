package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/mq"
	"XiaoLong-Ridy/rpc/paysvc/internal/model"
	"XiaoLong-Ridy/rpc/paysvc/internal/repository"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// 支付宝交易成功状态
const alipayTradeSuccess = "TRADE_SUCCESS"

// 支付回调阶段的业务错误（用于和底层错误区分，方便上层按错误码重试）。
var (
	// 回调金额与支付单金额不一致：可能被篡改/错发，拒绝处理。
	ErrPaymentAmountMismatch = errors.New("notify total amount does not match payment order")
	// 支付单状态非法（非待支付）：回调已处理或回调早于正常流程。
	ErrPaymentInvalidStatus = errors.New("payment status invalid for notify")
)

// NotifyPaymentLogic 支付回调逻辑：
//   - 验签失败直接拒绝；
//   - 事务内完成"金额比对 + 状态校验 + 条件更新"，从根上防重放/防并发覆盖；
//   - 事务提交后再发 Kafka 事件（outbox-lite 模式：DB 优先、消息投递失败仅记日志，不阻塞响应）。
type NotifyPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNotifyPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NotifyPaymentLogic {
	return &NotifyPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// notifyResult 事务回调写入，事务外继续使用。
type notifyResult struct {
	statusChanged bool                // 是否在本事务内把状态从「待支付」流转为「支付成功」
	paymentId     int64               // 支付单 ID（用于更新 event_sent）
	paidAt        *time.Time          // 流水支付时间（可能为 nil）
	orderId       int64               // 支付单归属的订单 ID（用于发事件）
}

// NotifyPayment 支付回调主流程。
func (l *NotifyPaymentLogic) NotifyPayment(in *proto.NotifyPaymentRequest) (*proto.NotifyPaymentResponse, error) {
	// 1. 验签（防伪造）。
	if err := l.svcCtx.Verifier.Verify(l.ctx, in.NotifyRaw); err != nil {
		return nil, fmt.Errorf("verify sign failed: %w", err)
	}

	// 2. 仅处理支付成功状态，其余状态（TRADE_CLOSED 等）认为是退款通知，直接 ACK。
	if in.TradeStatus != alipayTradeSuccess {
		return &proto.NotifyPaymentResponse{Success: true, Message: "ignore non-success trade status"}, nil
	}

	// 3. 在事务内完成"读取-校验-条件更新"。
	// 事务提交后再异步发事件，避免消息投递把回调响应拖到支付宝 5 秒超时。
	var res notifyResult
	err := l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		// 读取支付单（不加行锁，下面用条件更新做"乐观锁"防并发覆盖）。
		var p model.Payment
		if err := tx.Where("payment_no = ?", in.PaymentNo).First(&p).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}
		res.paymentId = int64(p.Id)
		res.orderId = int64(p.OrderId)

		// 幂等：已支付直接视为成功，事务回滚不影响（仅读）。
		if p.Status == model.PaymentStatusPaid {
			res.statusChanged = false
			res.paidAt = p.PaidAt
			return nil
		}
		// 仅「待支付」状态可流转为「支付成功」。
		if p.Status != model.PaymentStatusPending {
			return ErrPaymentInvalidStatus
		}

		// 4. 金额比对（M5-1）：回调 total_amount_cents 必须等于支付单金额（分）。
		// 支付单金额已用 int64 分存储，无需 YuanToCents 转换，彻底避免 double 精度误差。
		if in.TotalAmountCents != p.AmountCents {
			return fmt.Errorf("%w: want=%d, got=%d", ErrPaymentAmountMismatch, p.AmountCents, in.TotalAmountCents)
		}

		// 5. 条件更新（M5-4）：WHERE id=? AND status=待支付
		// 仅当状态仍为「待支付」时才流转，杜绝并发场景下"读到旧值 → 被其他线程更新过 → 又被本线程覆盖"。
		var paidAt *time.Time
		if in.PaidAt > 0 {
			t := time.Unix(in.PaidAt, 0)
			paidAt = &t
		}
		// event_sent 在事务中先置为 false；Kafka 发送成功后再更新为 true，发送失败可依赖对账补发。
		updates := map[string]interface{}{
			"status":         model.PaymentStatusPaid,
			"transaction_id": in.TransactionId,
			"paid_at":        paidAt,
			"event_sent":     false,
		}
		txRes := tx.Model(&model.Payment{}).
			Where("id = ? AND status = ?", p.Id, model.PaymentStatusPending).
			Updates(updates)
		if txRes.Error != nil {
			return txRes.Error
		}
		if txRes.RowsAffected == 0 {
			// 并发竞争：其它回调已经把状态改了，回滚事务并判为幂等成功。
			res.statusChanged = false
			return nil
		}

		res.statusChanged = true
		res.paidAt = paidAt
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			return nil, err
		}
		if errors.Is(err, ErrPaymentInvalidStatus) {
			return nil, err
		}
		return nil, err
	}

	// 6. 仅当本事务确实把状态从「待支付」流转为「支付成功」时，才发 Kafka 事件。
	// 幂等/并发路径不发，避免重复事件。
	// 投递失败先本地重试；最终失败时 DB 已提交为「已支付/事件未发送」，由对账任务补发，
	// 仍返回 success 给支付宝，避免支付宝重复回调导致幂等分支无法补发事件。
	if res.statusChanged {
		if err := l.publishPaidEvent(in, in.TotalAmountCents, res.orderId); err != nil {
			l.Errorf("publish order.paid event failed, will be reconciled later: %v", err)
		} else if err := repository.NewPaymentRepo(l.svcCtx.DB).UpdateSelective(l.ctx, uint64(res.paymentId), map[string]interface{}{
			"event_sent": true,
		}); err != nil {
			l.Errorf("mark event_sent=true failed: %v", err)
		}
	}

	return &proto.NotifyPaymentResponse{Success: true, Message: "success"}, nil
}

// publishPaidEvent 发送支付成功事件到 Kafka，带有限重试；最终仍失败返回错误，
// 由调用方记日志并交给对账任务补发。
func (l *NotifyPaymentLogic) publishPaidEvent(in *proto.NotifyPaymentRequest, totalCents int64, orderId int64) error {
	event := &mq.OrderPaidEvent{
		OrderId:     orderId,
		PaymentNo:   in.PaymentNo,
		AmountCents: totalCents,
		PaidAt:      in.PaidAt,
	}
	data, err := mq.EncodeOrderPaidEvent(event)
	if err != nil {
		return err
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			// 简单退避：100ms、200ms
			time.Sleep(time.Duration(i) * 100 * time.Millisecond)
		}
		lastErr = l.svcCtx.Producer.Send(constants.TopicOrderPaid, in.PaymentNo, data)
		if lastErr == nil {
			return nil
		}
		l.Errorf("publish order.paid event attempt %d failed: %v", i+1, lastErr)
	}
	return fmt.Errorf("publish order.paid event failed after retries: %w", lastErr)
}
