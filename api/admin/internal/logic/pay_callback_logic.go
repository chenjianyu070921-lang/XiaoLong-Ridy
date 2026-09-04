package logic

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// PayCallbackLogic 处理支付宝异步通知。
//   - 解析表单（application/x-www-form-urlencoded）；
//   - 透传给 paysvc.NotifyPayment gRPC（验签、状态机、事务都在 paysvc 完成）；
//   - RPC 返回 success=true → 返回 "success"，否则 → 返回 "fail"（触发支付宝重试）。
type PayCallbackLogic struct {
	ctx *svc.ServiceContext
	logx.Logger
}

func NewPayCallbackLogic(ctx *svc.ServiceContext) *PayCallbackLogic {
	return &PayCallbackLogic{
		ctx:    ctx,
		Logger: logx.WithContext(context.Background()),
	}
}

// HandleAlipayNotify 返回支付宝要求的字符串响应："success" 或 "fail"。
func (l *PayCallbackLogic) HandleAlipayNotify(req *http.Request) (string, error) {
	if err := req.ParseForm(); err != nil {
		return "fail", err
	}

	paymentNo := req.FormValue("out_trade_no")
	if paymentNo == "" {
		return "fail", errors.New("missing out_trade_no")
	}

	amountCents := int64(0)
	if amt := req.FormValue("total_amount"); amt != "" {
		if f, err := strconv.ParseFloat(amt, 64); err == nil {
			amountCents = priceutil.YuanToCents(f)
		}
	}

	paidAt := int64(0)
	if t := req.FormValue("gmt_payment"); t != "" {
		if pt, err := time.Parse("2006-01-02 15:04:05", t); err == nil {
			paidAt = pt.Unix()
		}
	}

	raw := buildNotifyRaw(req.Form)

	if l.ctx.PaySvc == nil {
		return "fail", errors.New("pay client not configured")
	}
	resp, err := l.ctx.PaySvc.NotifyPayment(req.Context(), &proto.NotifyPaymentRequest{
		PaymentNo:        paymentNo,
		TradeStatus:      req.FormValue("trade_status"),
		TransactionId:    req.FormValue("trade_no"),
		TotalAmountCents: amountCents,
		PaidAt:           paidAt,
		NotifyRaw:        raw,
	})
	if err != nil || resp == nil || !resp.Success {
		return "fail", err
	}
	return "success", nil
}

// buildNotifyRaw 把 url.Values 序列化为按 key 排序的 "k1=v1&k2=v2" 形式（支付宝验签规范）。
func buildNotifyRaw(form url.Values) string {
	if len(form) == 0 {
		return ""
	}
	keys := make([]string, 0, len(form))
	for k := range form {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('&')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		if len(form[k]) > 0 {
			sb.WriteString(form[k][0])
		}
	}
	return sb.String()
}
