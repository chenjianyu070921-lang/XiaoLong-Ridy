package types

// ListReviewsRequest 是分页查询司机评价的请求参数。
type ListReviewsRequest struct {
	Page     int32 `form:"page"`
	PageSize int32 `form:"pageSize"`
}

// ReviewItem 表示一条评价记录，Direction 区分收到/给出的评价。
type ReviewItem struct {
	OrderID   int64  `json:"orderId"`
	UserID    int64  `json:"userId"`
	DriverID  int64  `json:"driverId"`
	Rating    int32  `json:"rating"`
	Comment   string `json:"comment"`
	Tags      string `json:"tags"`
	CreatedAt int64  `json:"createdAt"`
	Direction string `json:"direction"`
}

// ListReceivedReviewsResponse 返回司机收到的乘客评价列表。
type ListReceivedReviewsResponse struct {
	List     []ReviewItem `json:"list"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"pageSize"`
}

// ListGivenReviewsRequest 是分页查询司机给出评价的请求参数。
type ListGivenReviewsRequest struct {
	Page     int32 `form:"page"`
	PageSize int32 `form:"pageSize"`
}

// ListGivenReviewsResponse 返回司机给出的评价列表。
type ListGivenReviewsResponse struct {
	List     []ReviewItem `json:"list"`
	Total    int64        `json:"total"`
	Page     int32        `json:"page"`
	PageSize int32        `json:"pageSize"`
}

// SubmitDriverReviewRequest 表示司机评价乘客的请求参数。
type SubmitDriverReviewRequest struct {
	OrderID int64  `json:"orderId"`
	Rating  int32  `json:"rating"`
	Comment string `json:"comment"`
	Tags    string `json:"tags"`
}

// SubmitDriverReviewResponse 表示司机评价提交成功后的响应。
type SubmitDriverReviewResponse struct {
	ReviewID  int64 `json:"reviewId"`
	OrderID   int64 `json:"orderId"`
	DriverID  int64 `json:"driverId"`
	Rating    int32 `json:"rating"`
	CreatedAt int64 `json:"createdAt"`
}
