package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/internal/logic"
	"XiaoLong-Ridy/rpc/paysvc/internal/svc"
	"XiaoLong-Ridy/rpc/paysvc/proto"
)

// AlipayCallback 处理支付宝异步通知回调。
// 支付宝以 application/x-www-form-urlencoded 推送通知，这里把全部参数透传给
// NotifyPayment RPC 做验签与业务处理，并返回支付宝要求的 "success" / "fail"。
func AlipayCallback(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeFail(w)
			return
		}

		if err := r.ParseForm(); err != nil {
			writeFail(w)
			return
		}

		raw := make(map[string]string, len(r.Form))
		for k, v := range r.Form {
			if len(v) > 0 {
				raw[k] = v[0]
			}
		}

		req := buildNotifyRequest(raw)
		resp, err := logic.NewNotifyPaymentLogic(r.Context(), svcCtx).NotifyPayment(req)
		if err != nil || resp == nil || !resp.Success {
			writeFail(w)
			return
		}
		writeSuccess(w)
	}
}

func buildNotifyRequest(raw map[string]string) *proto.NotifyPaymentRequest {
	req := &proto.NotifyPaymentRequest{}

	// 优先使用本系统自定义字段；兼容支付宝标准字段。
	if v := raw["payment_no"]; v != "" {
		req.PaymentNo = v
	} else if v = raw["out_trade_no"]; v != "" {
		req.PaymentNo = v
	}

	if v := raw["transaction_id"]; v != "" {
		req.TransactionId = v
	} else if v = raw["trade_no"]; v != "" {
		req.TransactionId = v
	}

	// 金额：支付宝以元为单位字符串返回，转换为分。
	amountStr := raw["total_amount"]
	if amountStr == "" {
		amountStr = raw["totalAmount"]
	}
	if amountStr != "" {
		if f, err := strconv.ParseFloat(amountStr, 64); err == nil {
			req.TotalAmountCents = priceutil.YuanToCents(f)
		}
	}

	// 时间：优先 Unix 秒，其次支付宝 gmt_payment 格式。
	if v := raw["paid_at"]; v != "" {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.PaidAt = sec
		}
	} else if v = raw["gmt_payment"]; v != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			req.PaidAt = t.Unix()
		}
	}

	return req
}

func writeSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}

func writeFail(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK) // 支付宝要求返回 200，但 body 为 fail 会触发重试
	_, _ = w.Write([]byte("fail"))
}

// ParseNotifyForm 解析支付宝原始通知字符串（k1=v1&k2=v2 形式）。
func ParseNotifyForm(body string) map[string]string {
	raw := make(map[string]string)
	if body == "" {
		return raw
	}
	for _, pair := range strings.Split(body, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		raw[kv[0]] = kv[1]
	}
	return raw
}
