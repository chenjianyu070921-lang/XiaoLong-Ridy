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
//
// 关于回调验签：
//   - paysvc 在密钥齐全时启动强校验（M5-3），NotifyPayment 必须通过 RSA2 验签才会流转。
//   - 本脚本从 rpc/paysvc/etc/paysvc.yaml 读取 alipay.privateKey/appId，
//     用支付宝 Rsa2（SHA256withRSA）对回调参数做真实签名，构造合法的 NotifyRaw。
//   - 也可用环境变量 ALIPAY_PRIVATE_KEY / ALIPAY_APP_ID 覆盖 yaml（与 M5-8 语义一致）。
package main

import (
	"context"
	"crypto"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/paysvc/pay"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/smartwalle/alipay/v3"
	"github.com/smartwalle/ncrypto"
	"github.com/smartwalle/nsign"
	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/yaml.v2"
)

var (
	target   = flag.String("target", "127.0.0.1:50054", "paysvc 地址")
	orderID  = flag.Int64("order", 1001, "订单ID")
	userID   = flag.Int64("user", 2001, "用户ID")
	amount   = flag.Int64("amount", 2500, "支付金额（分）")
	refund   = flag.Int64("refund", 1000, "退款金额（分）")
	config   = flag.String("config", "rpc/paysvc/etc/paysvc.yaml", "paysvc 配置文件路径")
)

func main() {
	flag.Parse()

	client := pay.NewPay(zrpc.MustNewClient(zrpc.RpcClientConf{Target: *target}))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 读取支付宝私钥（用于构造合法回调签名）
	appID, privateKey := loadAlipayKeys(*config)

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

	// 2. 构造合法签名的支付宝支付成功回调
	next("NotifyPayment 模拟支付成功回调（RSA2 真实签名）")
	notifyRaw := buildNotifyRaw(appID, paymentNo, *amount, privateKey)
	notifyResp, err := client.NotifyPayment(ctx, &pay.NotifyPaymentRequest{
		PaymentNo:       paymentNo,
		TradeStatus:     "TRADE_SUCCESS",
		TransactionId:   "tx_" + paymentNo,
		TotalAmountCents: *amount,
		PaidAt:          time.Now().Unix(),
		NotifyRaw:       notifyRaw,
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

// loadAlipayKeys 读取支付宝 appId 与私钥：环境变量优先，其次 yaml（M5-8 同款语义）。
func loadAlipayKeys(cfgPath string) (appID, privateKey string) {
	if v := strings.TrimSpace(os.Getenv("ALIPAY_APP_ID")); v != "" {
		appID = v
	}
	if v := strings.TrimSpace(os.Getenv("ALIPAY_PRIVATE_KEY")); v != "" {
		privateKey = v
	}
	if appID != "" && privateKey != "" {
		return appID, privateKey
	}

	type alipayCfg struct {
		AppId      string `yaml:"appId"`
		PrivateKey string `yaml:"privateKey"`
	}
	var cfg struct {
		Alipay alipayCfg `yaml:"alipay"`
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		log.Fatalf("读取配置文件 %s 失败: %v", cfgPath, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件 %s 失败: %v", cfgPath, err)
	}
	if appID == "" {
		appID = cfg.Alipay.AppId
	}
	if privateKey == "" {
		privateKey = cfg.Alipay.PrivateKey
	}
	return appID, privateKey
}

// buildNotifyRaw 构造支付宝回调参数串。
//   - 有私钥：用 RSA2 真实签名（对应生产模式 paysvc.yaml 启动的服务）；
//   - 无私钥：不签名直接返回（对应 dev/test 模式启动的服务，MockVerifier 不验签）。
func buildNotifyRaw(appID, paymentNo string, amountCents int64, privateKey string) string {
	params := url.Values{}
	params.Set("app_id", appID)
	params.Set("out_trade_no", paymentNo)
	params.Set("trade_no", "tx_"+paymentNo)
	params.Set("trade_status", "TRADE_SUCCESS")
	params.Set("total_amount", fmt.Sprintf("%.2f", float64(amountCents)/100))
	params.Set("notify_time", time.Now().Format("2006-01-02 15:04:05"))
	params.Set("notify_type", "trade_status_sync")

	if strings.TrimSpace(appID) == "" || strings.TrimSpace(privateKey) == "" {
		fmt.Println("  (alipay keys 为空，按 dev/test 模式处理：不签名，服务端 MockVerifier 放行)")
		return params.Encode()
	}

	sign, err := rsa2Sign(params, privateKey)
	if err != nil {
		log.Fatalf("生成回调签名失败: %v", err)
	}
	params.Set("sign", sign)
	params.Set("sign_type", "RSA2")
	return params.Encode()
}

// rsa2Sign 对待签名内容做 SHA256withRSA 签名（支付宝 RSA2）。
// 复用 alipay 包自定义 Encoder + nsign，与服务端 VerifySign 的拼接规则 100% 一致。
// privateKey 支持裸 base64（沙箱）或 PEM。
func rsa2Sign(params url.Values, privateKey string) (string, error) {
	// 1. 解析私钥：先 PKCS8，失败回退 PKCS1
	decoder := ncrypto.DecodePrivateKey([]byte(strings.TrimSpace(privateKey)))
	priv, err := decoder.PKCS8().RSAPrivateKey()
	if err != nil {
		priv, err = decoder.PKCS1().RSAPrivateKey()
	}
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}

	// 2. 与服务端 VerifySign 相同：alipay.Encoder{} + nsign + RSA2(SHA256)
	signer := nsign.New(
		nsign.WithMethod(nsign.NewRSAMethod(crypto.SHA256, priv, nil)),
		nsign.WithEncoder(alipay.Encoder{}),
	)
	sig, err := signer.SignValues(params, nsign.WithIgnore("sign", "sign_type"))
	if err != nil {
		return "", err
	}
	// SignValues 返回二进制签名，支付宝回调里 sign 是 base64 编码的字符串。
	return base64.StdEncoding.EncodeToString(sig), nil
}
