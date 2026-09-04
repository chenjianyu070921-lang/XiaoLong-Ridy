// Package demoapi 提供模块五答辩演示服务（rpc/pricesvc/demo）所需的支付侧适配接口。
//
// 为什么需要本包：
//
//	Go 的 internal 规则不允许 scripts/ 等位于 rpc/paysvc 之外的路径
//	直接 import rpc/paysvc/internal/...，因此这里把支付侧内部实现
//	（渠道 mock、结算规则）以导出函数的形式暴露给演示服务复用，
//	保证演示服务的支付数值与 paysvc 真实实现 100% 一致。
package demoapi

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/internal/channel"
	"XiaoLong-Ridy/rpc/paysvc/internal/rule"
)

// GenPaymentNo 生成平台支付单号，格式与 paysvc.CreatePayment.genPaymentNo 完全一致：
// PAY + 时间戳 + 微秒。
func GenPaymentNo() string {
	now := time.Now()
	return fmt.Sprintf("PAY%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
}

// GenRefundNo 生成退款单号，格式与 paysvc.RefundPayment.genRefundNo 完全一致：
// RF + 时间戳 + 微秒。
func GenRefundNo() string {
	now := time.Now()
	return fmt.Sprintf("RF%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
}

// CreateMockOrder 通过 paysvc 的 MockChannel 生成模拟支付参数，
// 返回的 transactionId / payParams 格式与真实 CreatePayment 一致。
func CreateMockOrder(paymentNo string, amountCents int64, channelName string) (transactionId, payParams string, err error) {
	ch := channel.NewMockChannel(channelName)
	res, err := ch.CreateOrder(context.Background(), paymentNo, amountCents)
	if err != nil {
		return "", "", err
	}
	return res.TransactionId, res.PayParams, nil
}

// CalcSettlement 计算司机结算（平台抽成 + 司机实收），复用 paysvc 结算规则。
func CalcSettlement(totalCents int64, commissionRate float64) (commissionCents, incomeCents int64) {
	return rule.CalcSettlement(totalCents, commissionRate)
}
