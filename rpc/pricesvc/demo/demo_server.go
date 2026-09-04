// 模块五答辩演示服务：以极简 HTTP 服务暴露计价/优惠/支付/结算全链路。
//
// 核心设计：计算全部复用模块五真实实现，保证演示数值与线上 100% 一致：
//   - 价格预估：internal/rule.Estimate（含夜间跨天/高峰判断，同 EstimatePriceLogic）
//   - 优惠抵扣：internal/rule.CalculateDiscount
//   - 计价规则：scripts/sql/migrate/07_price_rule_seed.sql 的北京三条真实规则
//   - 支付渠道/结算：通过 rpc/paysvc/demoapi 复用 paysvc 的 MockChannel 与结算规则
//     （Go internal 规则限制，scripts/ 无法直接 import rpc/.../internal/...）
//
// 用法：
//
//	go run ./rpc/pricesvc/demo -addr 127.0.0.1:8787
//
// 前端页面 docs/module5/demo/index.html 启动时会自动探测本服务：
// 连通 → 真实模式（本服务计算）；不通 → 页面自动降级为内置 JS 计算，保证答辩零依赖。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"XiaoLong-Ridy/common/priceutil"
	"XiaoLong-Ridy/rpc/paysvc/demoapi"
	"XiaoLong-Ridy/rpc/pricesvc/internal/model"
	"XiaoLong-Ridy/rpc/pricesvc/internal/rule"
)

// ────────────────────────── 计价规则（seed 真实数据）──────────────────────────

var beijingRules = []*model.PriceRule{
	{
		Id: 1, Name: "北京特惠快车", CityCode: "110000", CarType: 1,
		BasePrice: 8, BaseDistanceKm: 2, PerKmPrice: 1.80, PerMinutePrice: 0.40,
		NightStartTime: sptr("23:00:00"), NightEndTime: sptr("05:00:00"),
		NightSurcharge: 5, DynamicMaxFactor: 1.20, Status: 1,
	},
	{
		Id: 2, Name: "北京快车", CityCode: "110000", CarType: 2,
		BasePrice: 12, BaseDistanceKm: 3, PerKmPrice: 2.50, PerMinutePrice: 0.50,
		NightStartTime: sptr("23:00:00"), NightEndTime: sptr("05:00:00"),
		NightSurcharge: 8, DynamicMaxFactor: 1.50, Status: 1,
	},
	{
		Id: 3, Name: "北京拼车", CityCode: "110000", CarType: 3,
		BasePrice: 6, BaseDistanceKm: 2, PerKmPrice: 1.20, PerMinutePrice: 0.30,
		NightStartTime: sptr("23:00:00"), NightEndTime: sptr("05:00:00"),
		NightSurcharge: 3, DynamicMaxFactor: 1.10, Status: 1,
	},
}

func sptr(s string) *string { return &s }

// ────────────────────────── 支付状态机（内存）──────────────────────────

const (
	payStatusPending  = 1 // 待支付
	payStatusPaid     = 2 // 支付成功
	payStatusFailed   = 3 // 支付失败
	payStatusRefunded = 4 // 已退款
)

type paymentRecord struct {
	PaymentID      int64
	PaymentNo      string
	OrderID        int64
	UserID         int64
	AmountCents    int64
	Channel        string
	TransactionID  string
	PayParams      string
	Status         int32
	RefundedCents  int64
	NotifyCount    int    // 回调次数，用于演示幂等
	LastNotifyTime string // 最近一次回调时间
}

// ────────────────────────── 接口请求/响应结构 ──────────────────────────

type estimateReq struct {
	CityCode  string `json:"cityCode"`
	CarType   int    `json:"carType"`
	DistanceM int64  `json:"distanceM"`
	DurationS int64  `json:"durationS"`
	Timestamp int64  `json:"timestamp"`
}

type priceDetailResp struct {
	BaseFeeCents     int64 `json:"baseFeeCents"`
	DistanceFeeCents int64 `json:"distanceFeeCents"`
	TimeFeeCents     int64 `json:"timeFeeCents"`
	NightFeeCents    int64 `json:"nightFeeCents"`
	DynamicFeeCents  int64 `json:"dynamicFeeCents"`
	TotalCents       int64 `json:"totalCents"`
}

type ruleResp struct {
	Id               uint64  `json:"id"`
	Name             string  `json:"name"`
	CityCode         string  `json:"cityCode"`
	CarType          int8    `json:"carType"`
	BasePrice        float64 `json:"basePrice"`
	BaseDistanceKm   float64 `json:"baseDistanceKm"`
	PerKmPrice       float64 `json:"perKmPrice"`
	PerMinutePrice   float64 `json:"perMinutePrice"`
	NightStartTime   string  `json:"nightStartTime"`
	NightEndTime     string  `json:"nightEndTime"`
	NightSurcharge   float64 `json:"nightSurcharge"`
	DynamicMaxFactor float64 `json:"dynamicMaxFactor"`
}

type estimateResp struct {
	Rule       *ruleResp        `json:"rule"`
	Detail     *priceDetailResp `json:"detail"`
	TotalCents int64            `json:"totalCents"`
	IsNight    bool             `json:"isNight"`
	IsPeak     bool             `json:"isPeak"`
	Factor     float64          `json:"factor"`
}

type couponReq struct {
	Type             int32 `json:"type"` // 1固定金额 2折扣
	FaceValueCents   int64 `json:"faceValueCents"`
	Discount         int32 `json:"discount"` // 折扣券折扣，80 表示 8 折
	ThresholdCents   int64 `json:"thresholdCents"`
	MaxDiscountCents int64 `json:"maxDiscountCents"`
}

type discountReq struct {
	TotalCents int64     `json:"totalCents"`
	Coupon     couponReq `json:"coupon"`
}

type discountResp struct {
	DiscountAmountCents  int64 `json:"discountAmountCents"`
	PlatformSubsidyCents int64 `json:"platformSubsidyCents"`
	PayableAmountCents   int64 `json:"payableAmountCents"`
}

type createPaymentReq struct {
	OrderId     int64  `json:"orderId"`
	UserId      int64  `json:"userId"`
	AmountCents int64  `json:"amountCents"`
	Channel     string `json:"channel"` // wechat / alipay / balance
}

type createPaymentResp struct {
	PaymentId     int64  `json:"paymentId"`
	PaymentNo     string `json:"paymentNo"`
	TransactionId string `json:"transactionId"`
	PayParams     string `json:"payParams"`
	Status        int32  `json:"status"`
}

type notifyResp struct {
	Success    bool   `json:"success"`
	Idempotent bool   `json:"idempotent"` // true 表示命中重复回调幂等
	Message    string `json:"message"`
	PaymentNo  string `json:"paymentNo"`
	Status     int32  `json:"status"`
}

type getPaymentResp struct {
	PaymentNo         string `json:"paymentNo"`
	OrderId           int64  `json:"orderId"`
	AmountCents       int64  `json:"amountCents"`
	Channel           string `json:"channel"`
	Status            int32  `json:"status"`
	TransactionId     string `json:"transactionId"`
	RefundAmountCents int64  `json:"refundAmountCents"`
	NotifyCount       int    `json:"notifyCount"`
	LastNotifyTime    string `json:"lastNotifyTime"`
}

type refundReq struct {
	PaymentNo         string `json:"paymentNo"`
	RefundAmountCents int64  `json:"refundAmountCents"`
	Reason            string `json:"reason"`
}

type refundResp struct {
	Success             bool   `json:"success"`
	RefundNo            string `json:"refundNo"`
	RefundedAmountCents int64  `json:"refundedAmountCents"`
	FullRefund          bool   `json:"fullRefund"`
	Status              int32  `json:"status"`
	Message             string `json:"message"`
}

type settleReq struct {
	TotalCents     int64   `json:"totalCents"`
	CommissionRate float64 `json:"commissionRate"` // %，如 20 表示 20%
}

type settleResp struct {
	PlatformCommissionCents int64   `json:"platformCommissionCents"`
	DriverIncomeCents       int64   `json:"driverIncomeCents"`
	CommissionRate          float64 `json:"commissionRate"`
}

// ────────────────────────── 业务逻辑（复用真实 rule 包）──────────────────────────

// estimate 复用 pricesvc 的计价引擎与夜间/高峰判断，计算方式与
// EstimatePriceLogic.EstimatePrice 完全一致（含高峰 factor=规则 DynamicMaxFactor）。
func estimate(req estimateReq) (*estimateResp, error) {
	var pr *model.PriceRule
	for _, r := range beijingRules {
		if r.CityCode == req.CityCode && int(r.CarType) == req.CarType {
			pr = r
			break
		}
	}
	if pr == nil {
		return nil, fmt.Errorf("price rule not found: city_code=%s car_type=%d", req.CityCode, req.CarType)
	}

	now := time.Now()
	if req.Timestamp > 0 {
		now = time.Unix(req.Timestamp, 0)
	}
	nightStart, nightEnd := "", ""
	if pr.NightStartTime != nil {
		nightStart = *pr.NightStartTime
	}
	if pr.NightEndTime != nil {
		nightEnd = *pr.NightEndTime
	}
	isNight := rule.IsNightTime(now, nightStart, nightEnd)
	isPeak := rule.IsPeakTime(now)

	factor := 1.0
	if isPeak {
		factor = pr.DynamicMaxFactor
	}

	detail, err := rule.Estimate(rule.PriceRuleInput{
		BasePriceCents:      priceutil.YuanToCents(pr.BasePrice),
		BaseDistanceM:       int64(pr.BaseDistanceKm * 1000),
		PerKmPriceCents:     priceutil.YuanToCents(pr.PerKmPrice),
		PerMinutePriceCents: priceutil.YuanToCents(pr.PerMinutePrice),
		NightSurchargeCents: priceutil.YuanToCents(pr.NightSurcharge),
		DynamicMaxFactor:    pr.DynamicMaxFactor,
	}, req.DistanceM, req.DurationS, isNight, factor)
	if err != nil {
		return nil, err
	}

	return &estimateResp{
		Rule: &ruleResp{
			Id: pr.Id, Name: pr.Name, CityCode: pr.CityCode, CarType: pr.CarType,
			BasePrice: pr.BasePrice, BaseDistanceKm: pr.BaseDistanceKm,
			PerKmPrice: pr.PerKmPrice, PerMinutePrice: pr.PerMinutePrice,
			NightStartTime: nightStart, NightEndTime: nightEnd,
			NightSurcharge: pr.NightSurcharge, DynamicMaxFactor: pr.DynamicMaxFactor,
		},
		Detail: &priceDetailResp{
			BaseFeeCents:     detail.BaseFeeCents,
			DistanceFeeCents: detail.DistanceFeeCents,
			TimeFeeCents:     detail.TimeFeeCents,
			NightFeeCents:    detail.NightFeeCents,
			DynamicFeeCents:  detail.DynamicFeeCents,
			TotalCents:       detail.TotalCents,
		},
		TotalCents: detail.TotalCents,
		IsNight:    isNight,
		IsPeak:     isPeak,
		Factor:     factor,
	}, nil
}

// discount 复用 pricesvc 的优惠计算规则。
func discount(req discountReq) (*discountResp, error) {
	res, err := rule.CalculateDiscount(req.TotalCents, rule.CouponInput{
		Type:             req.Coupon.Type,
		FaceValueCents:   req.Coupon.FaceValueCents,
		Discount:         req.Coupon.Discount,
		ThresholdCents:   req.Coupon.ThresholdCents,
		MaxDiscountCents: req.Coupon.MaxDiscountCents,
	})
	if err != nil {
		return nil, err
	}
	return &discountResp{
		DiscountAmountCents:  res.DiscountAmountCents,
		PlatformSubsidyCents: res.PlatformSubsidyCents,
		PayableAmountCents:   res.PayableAmountCents,
	}, nil
}

// ────────────────────────── HTTP 服务 ──────────────────────────

type server struct {
	mu       sync.Mutex
	payments map[string]*paymentRecord // paymentNo -> record
	seq      int64
}

func newServer() *server {
	return &server{payments: make(map[string]*paymentRecord)}
}

func (s *server) withCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *server) writeJSON(w http.ResponseWriter, code int, v interface{}) {
	s.withCORS(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *server) writeErr(w http.ResponseWriter, code int, msg string) {
	s.writeJSON(w, code, map[string]interface{}{"error": msg})
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		s.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "mode": "real"})
	})

	mux.HandleFunc("GET /api/rules", func(w http.ResponseWriter, r *http.Request) {
		list := make([]*ruleResp, 0, len(beijingRules))
		for _, pr := range beijingRules {
			nightStart, nightEnd := "", ""
			if pr.NightStartTime != nil {
				nightStart = *pr.NightStartTime
			}
			if pr.NightEndTime != nil {
				nightEnd = *pr.NightEndTime
			}
			list = append(list, &ruleResp{
				Id: pr.Id, Name: pr.Name, CityCode: pr.CityCode, CarType: pr.CarType,
				BasePrice: pr.BasePrice, BaseDistanceKm: pr.BaseDistanceKm,
				PerKmPrice: pr.PerKmPrice, PerMinutePrice: pr.PerMinutePrice,
				NightStartTime: nightStart, NightEndTime: nightEnd,
				NightSurcharge: pr.NightSurcharge, DynamicMaxFactor: pr.DynamicMaxFactor,
			})
		}
		s.writeJSON(w, http.StatusOK, map[string]interface{}{"rules": list})
	})

	mux.HandleFunc("POST /api/estimate", func(w http.ResponseWriter, r *http.Request) {
		var req estimateReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}
		resp, err := estimate(req)
		if err != nil {
			s.writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("[estimate] car_type=%d distance=%dm duration=%ds -> total=%d分", req.CarType, req.DistanceM, req.DurationS, resp.TotalCents)
		s.writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("POST /api/discount", func(w http.ResponseWriter, r *http.Request) {
		var req discountReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}
		resp, err := discount(req)
		if err != nil {
			s.writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("[discount] total=%d分 -> 抵扣=%d分 实付=%d分", req.TotalCents, resp.DiscountAmountCents, resp.PayableAmountCents)
		s.writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("POST /api/payment/create", func(w http.ResponseWriter, r *http.Request) {
		var req createPaymentReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}
		if req.AmountCents <= 0 {
			s.writeErr(w, http.StatusBadRequest, "amount must be positive")
			return
		}
		if req.Channel == "" {
			req.Channel = "wechat"
		}

		paymentNo := demoapi.GenPaymentNo()
		transactionId, payParams, err := demoapi.CreateMockOrder(paymentNo, req.AmountCents, req.Channel)
		if err != nil {
			s.writeErr(w, http.StatusInternalServerError, "channel create order failed: "+err.Error())
			return
		}

		s.mu.Lock()
		s.seq++
		rec := &paymentRecord{
			PaymentID:     s.seq,
			PaymentNo:     paymentNo,
			OrderID:       req.OrderId,
			UserID:        req.UserId,
			AmountCents:   req.AmountCents,
			Channel:       req.Channel,
			TransactionID: transactionId,
			PayParams:     payParams,
			Status:        payStatusPending,
		}
		s.payments[paymentNo] = rec
		s.mu.Unlock()

		log.Printf("[payment/create] %s order=%d amount=%d分 channel=%s", paymentNo, req.OrderId, req.AmountCents, req.Channel)
		s.writeJSON(w, http.StatusOK, createPaymentResp{
			PaymentId:     rec.PaymentID,
			PaymentNo:     rec.PaymentNo,
			TransactionId: rec.TransactionID,
			PayParams:     rec.PayParams,
			Status:        rec.Status,
		})
	})

	mux.HandleFunc("POST /api/payment/notify", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PaymentNo string `json:"paymentNo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}

		s.mu.Lock()
		rec := s.payments[req.PaymentNo]
		if rec == nil {
			s.mu.Unlock()
			s.writeErr(w, http.StatusNotFound, "payment not found: "+req.PaymentNo)
			return
		}
		rec.NotifyCount++
		rec.LastNotifyTime = time.Now().Format("2006-01-02 15:04:05")
		// 幂等：状态已是支付成功，重复回调直接忽略。
		if rec.Status == payStatusPaid {
			status := rec.Status
			notifyCount := rec.NotifyCount
			s.mu.Unlock()
			log.Printf("[payment/notify] %s 重复回调，幂等忽略（第%d次）", req.PaymentNo, notifyCount)
			s.writeJSON(w, http.StatusOK, notifyResp{
				Success:    true,
				Idempotent: true,
				Message:    "重复回调已忽略（幂等：WHERE id=? AND status=待支付 条件更新命中 0 行）",
				PaymentNo:  req.PaymentNo,
				Status:     status,
			})
			return
		}
		if rec.Status != payStatusPending {
			status := rec.Status
			s.mu.Unlock()
			s.writeErr(w, http.StatusConflict, fmt.Sprintf("payment status=%d cannot transition", status))
			return
		}
		rec.Status = payStatusPaid
		rec.RefundedCents = 0
		status := rec.Status
		s.mu.Unlock()

		log.Printf("[payment/notify] %s 支付成功回调处理完成", req.PaymentNo)
		s.writeJSON(w, http.StatusOK, notifyResp{
			Success:    true,
			Idempotent: false,
			Message:    "支付成功（验签通过 → 金额比对一致 → 条件更新 status=2）",
			PaymentNo:  req.PaymentNo,
			Status:     status,
		})
	})

	mux.HandleFunc("GET /api/payment", func(w http.ResponseWriter, r *http.Request) {
		paymentNo := r.URL.Query().Get("paymentNo")
		s.mu.Lock()
		rec := s.payments[paymentNo]
		if rec == nil {
			s.mu.Unlock()
			s.writeErr(w, http.StatusNotFound, "payment not found: "+paymentNo)
			return
		}
		resp := getPaymentResp{
			PaymentNo:         rec.PaymentNo,
			OrderId:           rec.OrderID,
			AmountCents:       rec.AmountCents,
			Channel:           rec.Channel,
			Status:            rec.Status,
			TransactionId:     rec.TransactionID,
			RefundAmountCents: rec.RefundedCents,
			NotifyCount:       rec.NotifyCount,
			LastNotifyTime:    rec.LastNotifyTime,
		}
		s.mu.Unlock()
		s.writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("POST /api/payment/refund", func(w http.ResponseWriter, r *http.Request) {
		var req refundReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}

		s.mu.Lock()
		rec := s.payments[req.PaymentNo]
		if rec == nil {
			s.mu.Unlock()
			s.writeErr(w, http.StatusNotFound, "payment not found: "+req.PaymentNo)
			return
		}
		// 退款校验：仅支付成功可退、金额为正、不超过可退额。
		if rec.Status != payStatusPaid {
			s.mu.Unlock()
			s.writeErr(w, http.StatusConflict, "only paid payment can refund")
			return
		}
		if req.RefundAmountCents <= 0 {
			s.mu.Unlock()
			s.writeErr(w, http.StatusBadRequest, "refund amount must be positive")
			return
		}
		refundable := rec.AmountCents - rec.RefundedCents
		if req.RefundAmountCents > refundable {
			s.mu.Unlock()
			s.writeErr(w, http.StatusBadRequest, fmt.Sprintf("refund exceed refundable: %d", refundable))
			return
		}

		rec.RefundedCents += req.RefundAmountCents
		fullRefund := rec.RefundedCents >= rec.AmountCents
		if fullRefund {
			rec.Status = payStatusRefunded
		}
		resp := refundResp{
			Success:             true,
			RefundNo:            demoapi.GenRefundNo(),
			RefundedAmountCents: rec.RefundedCents,
			FullRefund:          fullRefund,
			Status:              rec.Status,
			Message:             "退款成功",
		}
		s.mu.Unlock()

		log.Printf("[payment/refund] %s 退款=%d分 累计=%d分 full=%v", req.PaymentNo, req.RefundAmountCents, resp.RefundedAmountCents, fullRefund)
		s.writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("POST /api/settle", func(w http.ResponseWriter, r *http.Request) {
		var req settleReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.writeErr(w, http.StatusBadRequest, "bad request: "+err.Error())
			return
		}
		commission, income := demoapi.CalcSettlement(req.TotalCents, req.CommissionRate)
		resp := settleResp{
			PlatformCommissionCents: commission,
			DriverIncomeCents:       income,
			CommissionRate:          req.CommissionRate,
		}
		log.Printf("[settle] total=%d分 rate=%g%% -> 平台=%d分 司机=%d分", req.TotalCents, req.CommissionRate, resp.PlatformCommissionCents, resp.DriverIncomeCents)
		s.writeJSON(w, http.StatusOK, resp)
	})

	// 静态文件托管：docs/module5/demo（浏览器直接打开 http://127.0.0.1:8787/ 即可演示）
	// 注意：http.FileServer 会把请求 /index.html 301 重定向到 /，而这里会把 / 映射到
	// /index.html，二者叠加会造成无限重定向（"too many redirects"），因此用 http.ServeFile
	// 按路径直接读文件，不做目录重定向。
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			s.withCORS(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// 非 API 路径统一走静态页面
		if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/health" {
			p := r.URL.Path
			if p == "/" {
				p = "/index.html"
			}
			if !strings.Contains(p, "..") {
				http.ServeFile(w, r, "docs/module5/demo"+p)
				return
			}
			http.NotFound(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "监听地址")
	flag.Parse()

	srv := newServer()
	log.Printf("模块五演示服务已启动：http://%s/", *addr)
	log.Printf("页面文件位于 docs/module5/demo/index.html，直接双击该文件也可（内置模拟模式）")

	if err := http.ListenAndServe(*addr, srv.handler()); err != nil {
		log.Fatalf("demo server exited: %v", err)
	}
}
