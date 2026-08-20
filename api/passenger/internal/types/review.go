package types

type SubmitReviewRequest struct {
	OrderID int64  `json:"orderId"`
	Rating  int32  `json:"rating"`
	Comment string `json:"comment"`
	Tags    string `json:"tags"`
}

type SubmitReviewResponse struct {
	ReviewID  int64 `json:"reviewId"`
	OrderID   int64 `json:"orderId"`
	DriverID  int64 `json:"driverId"`
	Rating    int32 `json:"rating"`
	CreatedAt int64 `json:"createdAt"`
}
