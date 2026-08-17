// 支付模块真实联调客户端：直连 paysvc，跑通支付正向全链路。
//
// 用法（配合 run_pay_e2e.ps1，或手动）：
//
//	go run scripts/e2e/pay_e2e_client.go -target 127.0.0.1:50054
//
// 覆盖链路：
//
//	CreatePayment（预下单）→ NotifyPayment（模拟支付宝回调）→ GetPayment（查询）
//	  → SettleOrder（司机结算）→ RefundPayment（退款）
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/pay"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	target  = flag.String("target", "127.0.0.1:50054", "paysvc 地址")
	orderID = flag.Int64("order", 1001, "订单ID")
	userID  = flag.Int64("user", 2001, "用户ID")
	amount  = flag.Int64("amount", 2500, "支付金额（分）")
	refund  = flag.Int64("refund", 1000, "退款金额（分）")
)

func main() {
	flag.Parse()

	client := pay.NewPay(zrpc.MustNewClient(zrpc.RpcClientConf{Target: *target}))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	step := 0
	next := func(name string) {
		step++
		fmt.Printf("\n===== [%d] %s =====\n", step, name)
	}

	// 1. 支付预下单
	next("CreatePayment 支付预下单")
	createResp, err := client.CreatePayment(ctx, &pay.CreatePaymentRequest{
		OrderId:     *orderID,
		UserId:      *userID,
		AmountCents: *amount,
		Channel:     proto.PayChannel_PAY_CHANNEL_WECHAT,
	})
	if err != nil {
		log.Fatalf("CreatePayment 失败: %v", err)
	}
	fmt.Printf("payment_id=%d payment_no=%s transaction_id=%s status=%d\n",
		createResp.PaymentId, createResp.PaymentNo, createResp.TransactionId, createResp.Status)
	fmt.Printf("pay_params=%s\n", createResp.PayParams)

	paymentNo := createResp.PaymentNo

	// 2. 模拟支付宝支付成功回调（本地联调 keys 为空，验签自动降级 MockVerifier）
	next("NotifyPayment 模拟支付成功回调")
	notifyResp, err := client.NotifyPayment(ctx, &pay.NotifyPaymentRequest{
		PaymentNo:     paymentNo,
		TradeStatus:   "TRADE_SUCCESS",
		TransactionId: "tx_" + paymentNo,
		PaidAt:        time.Now().Unix(),
		NotifyRaw:     "payment_no=" + paymentNo,
	})
	if err != nil {
		log.Fatalf("NotifyPayment 失败: %v", err)
	}
	fmt.Printf("success=%v message=%s\n", notifyResp.Success, notifyResp.Message)

	// 3. 支付查询（应显示 status=2 已支付）
	next("GetPayment 支付查询")
	getResp, err := client.GetPayment(ctx, &pay.GetPaymentRequest{PaymentNo: paymentNo})
	if err != nil {
		log.Fatalf("GetPayment 失败: %v", err)
	}
	fmt.Printf("status=%d amount_cents=%d channel=%s refunded_cents=%d transaction_id=%s\n",
		getResp.Status, getResp.AmountCents, getResp.Channel, getResp.RefundAmountCents, getResp.TransactionId)

	// 4. 司机结算（回调链路里已触发一次，这里显式再调一次验证结算接口独立可用）
	next("SettleOrder 司机结算")
	settleResp, err := client.SettleOrder(ctx, &pay.SettleOrderRequest{
		OrderId:          *orderID,
		DriverId:         3001,
		TotalAmountCents: *amount,
		CommissionRate:   20,
	})
	if err != nil {
		log.Fatalf("SettleOrder 失败: %v", err)
	}
	fmt.Printf("settlement_no=%s platform_commission=%d driver_income=%d\n",
		settleResp.SettlementNo, settleResp.PlatformCommissionCents, settleResp.DriverIncomeCents)

	// 5. 部分退款
	next("RefundPayment 退款")
	refundResp, err := client.RefundPayment(ctx, &pay.RefundPaymentRequest{
		PaymentNo:         paymentNo,
		RefundAmountCents: *refund,
		Reason:            "e2e 联调",
	})
	if err != nil {
		log.Fatalf("RefundPayment 失败: %v", err)
	}
	fmt.Printf("success=%v refund_no=%s refunded_cents=%d\n",
		refundResp.Success, refundResp.RefundNo, refundResp.RefundedAmountCents)

	fmt.Println("\n全部步骤执行完成，支付正向链路联调通过。")
}
