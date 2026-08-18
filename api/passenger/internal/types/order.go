package types

// CreateOrderRequest 表示乘客下单接口的请求参数。
type CreateOrderRequest struct {
	CarType                int32   `json:"carType"`
	FromAddress            string  `json:"fromAddress"`
	FromLongitude          float64 `json:"fromLongitude"`
	FromLatitude           float64 `json:"fromLatitude"`
	ToAddress              string  `json:"toAddress"`
	ToLongitude            float64 `json:"toLongitude"`
	ToLatitude             float64 `json:"toLatitude"`
	UserCouponID           uint64  `json:"userCouponId"`
	CouponID               int64   `json:"couponId"`
	CouponType             int32   `json:"couponType"`
	CouponFaceValueCents   int64   `json:"couponFaceValueCents"`
	CouponDiscount         int32   `json:"couponDiscount"`
	CouponThresholdCents   int64   `json:"couponThresholdCents"`
	CouponMaxDiscountCents int64   `json:"couponMaxDiscountCents"`
}

// CreateOrderResponse 表示乘客下单成功后的返回数据。
type CreateOrderResponse struct {
	OrderID             int64  `json:"orderId"`
	OrderNo             string `json:"orderNo"`
	EstimatedPriceCents int64  `json:"estimatedPriceCents"`
	OriginalPriceCents  int64  `json:"originalPriceCents"`
	DiscountAmountCents int64  `json:"discountAmountCents"`
	PayableAmountCents  int64  `json:"payableAmountCents"`
	UserCouponID        uint64 `json:"userCouponId"`
	Status              int32  `json:"status"`
	CreatedAt           int64  `json:"createdAt"`
}

// ListOrdersRequest 表示订单列表查询请求参数。
type ListOrdersRequest struct {
	Status   int32 `json:"status"`
	Page     int32 `json:"page"`
	PageSize int32 `json:"pageSize"`
}

// OrderSummary 表示订单列表中的单条摘要信息。
type OrderSummary struct {
	OrderID             int64  `json:"orderId"`
	OrderNo             string `json:"orderNo"`
	FromAddress         string `json:"fromAddress"`
	ToAddress           string `json:"toAddress"`
	Status              int32  `json:"status"`
	EstimatedPriceCents int64  `json:"estimatedPriceCents"`
	CreatedAt           int64  `json:"createdAt"`
}

// ListOrdersResponse 表示订单列表查询响应。
type ListOrdersResponse struct {
	List     []OrderSummary `json:"list"`
	Total    int64          `json:"total"`
	Page     int32          `json:"page"`
	PageSize int32          `json:"pageSize"`
}

// GetOrderRequest 表示订单详情查询请求参数。
type GetOrderRequest struct {
	OrderID int64 `json:"orderId"`
}

// OrderDetail 表示订单详情数据。
type OrderDetail struct {
	OrderID             int64   `json:"orderId"`
	OrderNo             string  `json:"orderNo"`
	UserID              int64   `json:"userId"`
	DriverID            int64   `json:"driverId"`
	CarType             int32   `json:"carType"`
	FromAddress         string  `json:"fromAddress"`
	FromLongitude       float64 `json:"fromLongitude"`
	FromLatitude        float64 `json:"fromLatitude"`
	ToAddress           string  `json:"toAddress"`
	ToLongitude         float64 `json:"toLongitude"`
	ToLatitude          float64 `json:"toLatitude"`
	EstimatedDistanceM  int64   `json:"estimatedDistanceM"`
	EstimatedDurationS  int64   `json:"estimatedDurationS"`
	EstimatedPriceCents int64   `json:"estimatedPriceCents"`
	Status              int32   `json:"status"`
	CancelReason        string  `json:"cancelReason"`
	CancelBy            string  `json:"cancelBy"`
	CreatedAt           int64   `json:"createdAt"`
	UpdatedAt           int64   `json:"updatedAt"`
}

// CancelOrderRequest 表示取消订单请求参数。
type CancelOrderRequest struct {
	OrderID int64  `json:"orderId"`
	Reason  string `json:"reason"`
}

// CancelOrderResponse 表示取消订单响应。
type CancelOrderResponse struct {
	OrderID int64 `json:"orderId"`
	Status  int32 `json:"status"`
}

// PayOrderRequest 表示乘客发起支付预下单的请求参数。
type PayOrderRequest struct {
	OrderID int64 `json:"orderId"`
	Channel int32 `json:"channel"`
}

// PayOrderResponse 表示 paysvc 创建支付单后的返回数据。
type PayOrderResponse struct {
	PaymentID     int64  `json:"paymentId"`
	PaymentNo     string `json:"paymentNo"`
	TransactionID string `json:"transactionId"`
	PayParams     string `json:"payParams"`
	Status        int32  `json:"status"`
}
