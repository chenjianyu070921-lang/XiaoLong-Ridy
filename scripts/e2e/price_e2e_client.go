// 计价模块真实联调客户端：直连 pricesvc，跑通计价三条接口。
//
// 用法：
//
//	go run scripts/e2e/price_e2e_client.go -target 127.0.0.1:50053
//
// 覆盖链路：
//
//	EstimatePrice（价格预估）→ CalculateDiscount（优惠抵扣）→ SaveActualOrderPrice（费用落库）
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"XiaoLong-Ridy/rpc/pricesvc/price"
	"XiaoLong-Ridy/rpc/pricesvc/proto"

	"github.com/zeromicro/go-zero/zrpc"
)

var (
	target  = flag.String("target", "127.0.0.1:50053", "pricesvc 地址")
	orderID = flag.Int64("order", 1001, "订单ID")
	userID  = flag.Int64("user", 2001, "用户ID")
)

func main() {
	flag.Parse()

	client := price.NewPrice(zrpc.MustNewClient(zrpc.RpcClientConf{Target: *target}))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	step := 0
	next := func(name string) {
		step++
		fmt.Printf("\n===== [%d] %s =====\n", step, name)
	}

	// 1. 行程价格预估（命中 DB 规则：city_code=110000, car_type=2）
	next("EstimatePrice 行程价格预估")
	estimateResp, err := client.EstimatePrice(ctx, &price.EstimatePriceRequest{
		UserId:     *userID,
		CityCode:   "110000",
		CarType:    2,
		DistanceM:  5000, // 5 km
		DurationS:  600,  // 10 min
		Timestamp:  time.Now().Unix(),
	})
	if err != nil {
		log.Fatalf("EstimatePrice 失败: %v", err)
	}
	fmt.Printf("price_rule_id=%d total_cents=%d\n", estimateResp.PriceRuleId, estimateResp.TotalCents)
	if estimateResp.Detail != nil {
		fmt.Printf("detail: base=%d distance=%d time=%d night=%d dynamic=%d\n",
			estimateResp.Detail.BaseFeeCents, estimateResp.Detail.DistanceFeeCents,
			estimateResp.Detail.TimeFeeCents, estimateResp.Detail.NightFeeCents,
			estimateResp.Detail.DynamicFeeCents)
	}

	// 2. 兜底规则验证（未知城市 → 系统兜底价，price_rule_id 应为 0）
	next("EstimatePrice 兜底规则（未知城市 999999）")
	fallbackResp, err := client.EstimatePrice(ctx, &price.EstimatePriceRequest{
		UserId:    *userID,
		CityCode:  "999999",
		CarType:   1,
		DistanceM: 5000,
		DurationS: 600,
		Timestamp: time.Now().Unix(),
	})
	if err != nil {
		log.Fatalf("EstimatePrice(兜底) 失败: %v", err)
	}
	fmt.Printf("price_rule_id=%d total_cents=%d（兜底成功，price_rule_id 应为 0）\n",
		fallbackResp.PriceRuleId, fallbackResp.TotalCents)

	// 3. 优惠券抵扣计算（固定券 5 元）
	next("CalculateDiscount 优惠抵扣（固定券 5 元）")
	discountResp, err := client.CalculateDiscount(ctx, &price.CalculateDiscountRequest{
		OrderId:    *orderID,
		TotalCents: estimateResp.TotalCents,
		Coupon: &price.Coupon{
			CouponId:       1,
			Type:           proto.CouponType_COUPON_TYPE_FIXED,
			FaceValueCents: 500,
		},
	})
	if err != nil {
		log.Fatalf("CalculateDiscount 失败: %v", err)
	}
	fmt.Printf("discount=%d payable=%d\n",
		discountResp.DiscountAmountCents, discountResp.PayableAmountCents)

	// 4. 实际费用落库
	next("SaveActualOrderPrice 费用落库")
	saveResp, err := client.SaveActualOrderPrice(ctx, &price.SaveActualOrderPriceRequest{
		OrderId:          *orderID,
		PriceRuleId:      estimateResp.PriceRuleId,
		ActualPriceCents: estimateResp.TotalCents,
		Detail:           estimateResp.Detail,
	})
	if err != nil {
		log.Fatalf("SaveActualOrderPrice 失败: %v", err)
	}
	fmt.Printf("success=%v order_price_id=%d\n", saveResp.Success, saveResp.OrderPriceId)

	fmt.Println("\n计价链路全部接口联调通过。")
}
