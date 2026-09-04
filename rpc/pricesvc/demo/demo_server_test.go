package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// timeAt 构造指定时刻的 Unix 秒（用于夜间/高峰判断）。
func timeAt(hour, minute int) int64 {
	return time.Date(2026, 8, 27, hour, minute, 0, 0, time.Local).Unix()
}

func TestEstimateDaytimeKuaiche(t *testing.T) {
	// 北京快车 白天 12.5km / 1800s：
	//   起步 12.00 + (12.5-3)*2.5 + 0.5*30 = 1200+2375+1500 = 5075 分
	resp, err := estimate(estimateReq{
		CityCode: "110000", CarType: 2, DistanceM: 12500, DurationS: 1800,
		Timestamp: timeAt(10, 0),
	})
	if err != nil {
		t.Fatalf("estimate error: %v", err)
	}
	if resp.TotalCents != 5075 {
		t.Errorf("total = %d, want 5075", resp.TotalCents)
	}
	if resp.Detail.BaseFeeCents != 1200 || resp.Detail.DistanceFeeCents != 2375 || resp.Detail.TimeFeeCents != 1500 {
		t.Errorf("detail mismatch: %+v", resp.Detail)
	}
	if resp.Detail.NightFeeCents != 0 || resp.Detail.DynamicFeeCents != 0 {
		t.Errorf("daytime should have no night/dynamic fee: %+v", resp.Detail)
	}
	if resp.IsNight || resp.IsPeak {
		t.Errorf("10:00 should be day & non-peak: night=%v peak=%v", resp.IsNight, resp.IsPeak)
	}
}

func TestEstimatePeakKuaiche(t *testing.T) {
	// 早高峰 08:00 快车：factor=DynamicMaxFactor(1.5)
	//   basic=5075，dynamic=round(5075*0.5)=2538，total=7613
	resp, err := estimate(estimateReq{
		CityCode: "110000", CarType: 2, DistanceM: 12500, DurationS: 1800,
		Timestamp: timeAt(8, 0),
	})
	if err != nil {
		t.Fatalf("estimate error: %v", err)
	}
	if !resp.IsPeak {
		t.Error("08:00 should be peak")
	}
	if resp.Factor != 1.5 {
		t.Errorf("factor = %v, want 1.5", resp.Factor)
	}
	if resp.Detail.DynamicFeeCents != 2538 {
		t.Errorf("dynamic = %d, want 2538", resp.Detail.DynamicFeeCents)
	}
	if resp.TotalCents != 7613 {
		t.Errorf("total = %d, want 7613", resp.TotalCents)
	}
}

func TestEstimateNightCrossDay(t *testing.T) {
	// 夜间 23:30 快车 5km/600s：起步1200 + 里程(5-3)*2.5=500 + 时长0.5*10=500 + 夜间800 = 3000
	resp, err := estimate(estimateReq{
		CityCode: "110000", CarType: 2, DistanceM: 5000, DurationS: 600,
		Timestamp: timeAt(23, 30),
	})
	if err != nil {
		t.Fatalf("estimate error: %v", err)
	}
	if !resp.IsNight {
		t.Error("23:30 should be night")
	}
	if resp.Detail.NightFeeCents != 800 {
		t.Errorf("night fee = %d, want 800", resp.Detail.NightFeeCents)
	}
	if resp.TotalCents != 3000 {
		t.Errorf("total = %d, want 3000", resp.TotalCents)
	}
}

func TestEstimateNightTeshu(t *testing.T) {
	// 特惠快车 夜间 23:30 5km/600s：起步800 + 里程(5-2)*1.8=540 + 时长0.4*10=400 + 夜间500 = 2240
	resp, err := estimate(estimateReq{
		CityCode: "110000", CarType: 1, DistanceM: 5000, DurationS: 600,
		Timestamp: timeAt(23, 30),
	})
	if err != nil {
		t.Fatalf("estimate error: %v", err)
	}
	if resp.Detail.NightFeeCents != 500 {
		t.Errorf("night fee = %d, want 500", resp.Detail.NightFeeCents)
	}
	if resp.TotalCents != 2240 {
		t.Errorf("total = %d, want 2240", resp.TotalCents)
	}
}

func TestDiscountFixed(t *testing.T) {
	// 5075 分，5 元固定券 → 抵扣 500，实付 4575
	resp, err := discount(discountReq{
		TotalCents: 5075,
		Coupon:     couponReq{Type: 1, FaceValueCents: 500},
	})
	if err != nil {
		t.Fatalf("discount error: %v", err)
	}
	if resp.DiscountAmountCents != 500 || resp.PayableAmountCents != 4575 {
		t.Errorf("fixed coupon result mismatch: %+v", resp)
	}
}

func TestDiscountEightFold(t *testing.T) {
	// 5075 分，8 折券（满 2000，最大优惠 1000）：
	//   理论优惠 5075*20%=1015，受 MaxDiscountCents=1000 上限约束 → 抵扣 1000，实付 4075
	resp, err := discount(discountReq{
		TotalCents: 5075,
		Coupon:     couponReq{Type: 2, Discount: 80, ThresholdCents: 2000, MaxDiscountCents: 1000},
	})
	if err != nil {
		t.Fatalf("discount error: %v", err)
	}
	if resp.DiscountAmountCents != 1000 || resp.PayableAmountCents != 4075 {
		t.Errorf("eightfold coupon result mismatch: %+v", resp)
	}
}

func TestDiscountNotMeetThreshold(t *testing.T) {
	_, err := discount(discountReq{
		TotalCents: 1500,
		Coupon:     couponReq{Type: 2, Discount: 80, ThresholdCents: 2000},
	})
	if err == nil {
		t.Fatal("expected threshold error")
	}
}

// ────────────────────────── HTTP 全流程测试 ──────────────────────────

func doJSON(t *testing.T, s *server, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.handler().ServeHTTP(rec, req)
	return rec
}

func decodeResp(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("decode resp: %v, body=%s", err, rec.Body.String())
	}
}

func TestPaymentFlowHTTP(t *testing.T) {
	s := newServer()

	// 1. 创建支付单
	createRec := doJSON(t, s, http.MethodPost, "/api/payment/create", createPaymentReq{
		OrderId: 1001, UserId: 2001, AmountCents: 2500, Channel: "alipay",
	})
	var created createPaymentResp
	decodeResp(t, createRec, &created)
	if created.Status != payStatusPending {
		t.Errorf("create status = %d, want %d", created.Status, payStatusPending)
	}
	if created.PaymentNo == "" || created.TransactionId == "" || created.PayParams == "" {
		t.Errorf("create resp incomplete: %+v", created)
	}
	paymentNo := created.PaymentNo

	// 2. 支付成功回调
	notifyRec := doJSON(t, s, http.MethodPost, "/api/payment/notify", map[string]string{"paymentNo": paymentNo})
	var notified notifyResp
	decodeResp(t, notifyRec, &notified)
	if !notified.Success || notified.Idempotent || notified.Status != payStatusPaid {
		t.Errorf("notify resp mismatch: %+v", notified)
	}

	// 3. 重复回调 → 幂等
	dupRec := doJSON(t, s, http.MethodPost, "/api/payment/notify", map[string]string{"paymentNo": paymentNo})
	var dup notifyResp
	decodeResp(t, dupRec, &dup)
	if !dup.Success || !dup.Idempotent {
		t.Errorf("duplicate notify should be idempotent: %+v", dup)
	}

	// 4. 查询
	queryRec := doJSON(t, s, http.MethodGet, "/api/payment?paymentNo="+paymentNo, nil)
	var got getPaymentResp
	decodeResp(t, queryRec, &got)
	if got.Status != payStatusPaid || got.NotifyCount != 2 {
		t.Errorf("query mismatch: %+v", got)
	}

	// 5. 全额退款
	refundRec := doJSON(t, s, http.MethodPost, "/api/payment/refund", refundReq{
		PaymentNo: paymentNo, RefundAmountCents: 2500, Reason: "演示退款",
	})
	var rf refundResp
	decodeResp(t, refundRec, &rf)
	if !rf.Success || !rf.FullRefund || rf.Status != payStatusRefunded {
		t.Errorf("refund resp mismatch: %+v", rf)
	}

	// 6. 已退款后再退 → 拒绝
	againRec := doJSON(t, s, http.MethodPost, "/api/payment/refund", refundReq{
		PaymentNo: paymentNo, RefundAmountCents: 100, Reason: "再退",
	})
	if againRec.Code != http.StatusConflict {
		t.Errorf("refund after full refund should conflict, code=%d", againRec.Code)
	}
}

func TestHealth(t *testing.T) {
	s := newServer()
	rec := doJSON(t, s, http.MethodGet, "/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d", rec.Code)
	}
}
