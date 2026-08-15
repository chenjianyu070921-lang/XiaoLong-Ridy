package client

import (
	"context"
	"math"
)

const (
	// CouponTypeFixed 表示固定金额抵扣券。
	CouponTypeFixed int32 = 1
	// CouponTypeDiscount 表示折扣券，Discount 字段按百分比保存。
	CouponTypeDiscount int32 = 2
)

// EstimatePriceRequest 表示预估价格所需的核心参数。
type EstimatePriceRequest struct {
	UserID          int64
	CityCode        string
	CarType         int32
	FromLongitude   float64
	FromLatitude    float64
	ToLongitude     float64
	ToLatitude      float64
	EstimatedMeters int64
	EstimatedSecond int64
	Timestamp       int64
}

// EstimatePriceResponse 返回预估价格、里程和时长。
type EstimatePriceResponse struct {
	EstimatedPriceCents int64
	EstimatedDistanceM  int64
	EstimatedDurationS  int64
}

// Coupon 表示 passenger 传入的优惠券抵扣参数。
type Coupon struct {
	CouponID         int64
	Type             int32
	FaceValueCents   int64
	Discount         int32
	ThresholdCents   int64
	MaxDiscountCents int64
}

// CalculateDiscountRequest 表示优惠抵扣计算请求。
type CalculateDiscountRequest struct {
	OrderID    int64
	TotalCents int64
	Coupon     Coupon
}

// CalculateDiscountResponse 表示优惠抵扣计算结果。
type CalculateDiscountResponse struct {
	DiscountAmountCents  int64
	PlatformSubsidyCents int64
	PayableAmountCents   int64
}

// LocalClient 是本地开发和测试使用的价格服务实现。
type LocalClient struct {
}

// NewLocalClient 创建本地价格服务实现。
func NewLocalClient() *LocalClient {
	return &LocalClient{}
}

// EstimatePrice 根据起终点坐标计算一个稳定可复现的预估价格。
func (c *LocalClient) EstimatePrice(_ context.Context, req *EstimatePriceRequest) (*EstimatePriceResponse, error) {
	distanceM := req.EstimatedMeters
	if distanceM <= 0 {
		distanceM = haversineMeters(req.FromLatitude, req.FromLongitude, req.ToLatitude, req.ToLongitude)
	}
	if distanceM <= 0 {
		distanceM = 1000
	}

	durationS := req.EstimatedSecond
	if durationS <= 0 {
		durationS = int64(math.Ceil(float64(distanceM) / 250.0))
	}
	if durationS <= 0 {
		durationS = 60
	}

	base := int64(300)
	perKm := int64(180)
	switch req.CarType {
	case 1:
		base = 200
		perKm = 140
	case 2:
		base = 300
		perKm = 180
	case 3:
		base = 500
		perKm = 260
	}
	priceCents := base + int64(math.Ceil(float64(distanceM)/1000.0))*perKm + int64(math.Ceil(float64(durationS)/60.0))*10
	if priceCents < 0 {
		priceCents = 0
	}
	return &EstimatePriceResponse{
		EstimatedPriceCents: priceCents,
		EstimatedDistanceM:  distanceM,
		EstimatedDurationS:  durationS,
	}, nil
}

// CalculateDiscount 根据传入优惠券计算抵扣金额，供本地联调模拟 pricesvc 行为。
func (c *LocalClient) CalculateDiscount(_ context.Context, req *CalculateDiscountRequest) (*CalculateDiscountResponse, error) {
	total := req.TotalCents
	if total < 0 {
		total = 0
	}
	discount := int64(0)
	if req.Coupon.CouponID > 0 && total >= req.Coupon.ThresholdCents {
		switch req.Coupon.Type {
		case CouponTypeFixed:
			discount = req.Coupon.FaceValueCents
		case CouponTypeDiscount:
			if req.Coupon.Discount > 0 && req.Coupon.Discount < 100 {
				discount = total * int64(100-req.Coupon.Discount) / 100
				if req.Coupon.MaxDiscountCents > 0 && discount > req.Coupon.MaxDiscountCents {
					discount = req.Coupon.MaxDiscountCents
				}
			}
		}
	}
	if discount < 0 {
		discount = 0
	}
	if discount > total {
		discount = total
	}
	return &CalculateDiscountResponse{
		DiscountAmountCents: discount,
		PayableAmountCents:  total - discount,
	}, nil
}

// haversineMeters 使用球面距离公式估算两组经纬度之间的直线距离。
func haversineMeters(lat1, lon1, lat2, lon2 float64) int64 {
	const earthRadius = 6371000.0
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return int64(math.Round(earthRadius * c))
}
