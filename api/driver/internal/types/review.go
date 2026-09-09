package types

// ListReviewsRequest 是分页查询司机评价的请求参数。
type ListReviewsRequest struct {
	Page     int32 `form:"page"`
	PageSize int32 `form:"pageSize"`
}

// ReviewItem 表示一条乘客对司机的评价记录。
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

// ReviewSummaryResponse 返回司机评价概览：能否接收评价、完成单数、评价统计
// 与由 order_review 表聚合计算的服务平均分。
type ReviewSummaryResponse struct {
	CanReceiveReview    bool    `json:"canReceiveReview"`
	CompletedOrderCount int64   `json:"completedOrderCount"`
	ReviewCount         int64   `json:"reviewCount"`
	AvgRating           float64 `json:"avgRating"`
	ServiceScore        float64 `json:"serviceScore"`
}
