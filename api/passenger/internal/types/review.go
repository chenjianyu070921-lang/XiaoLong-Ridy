package types

// SubmitReviewRequest 表示乘客提交订单评价的请求参数。
type SubmitReviewRequest struct {
	OrderID int64  `json:"orderId"`
	Rating  int32  `json:"rating"`
	Comment string `json:"comment"`
	Tags    string `json:"tags"`
}

// SubmitReviewResponse 表示评价提交成功后的响应。
type SubmitReviewResponse struct {
	ReviewID  int64 `json:"reviewId"`
	OrderID   int64 `json:"orderId"`
	DriverID  int64 `json:"driverId"`
	Rating    int32 `json:"rating"`
	CreatedAt int64 `json:"createdAt"`
}
