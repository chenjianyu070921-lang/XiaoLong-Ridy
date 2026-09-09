package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// ErrReviewRepositoryNotConfigured 表示评价仓储未配置（MySQL 未初始化时降级）。
var ErrReviewRepositoryNotConfigured = errors.New("review repository not configured")

// ReviewLogic 封装司机端评价业务逻辑。
type ReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewReviewLogic 创建司机端评价逻辑处理器。
func NewReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLogic {
	return &ReviewLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListReceivedReviews 返回当前司机收到的乘客评价。
func (l *ReviewLogic) ListReceivedReviews(driverID int64, req *types.ListReviewsRequest) (*types.ListReceivedReviewsResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	page, pageSize := normalizeReviewPage(req.Page, req.PageSize)
	repo, err := l.reviewRepository()
	if err != nil {
		return nil, err
	}
	rows, total, err := repo.ListPassengerReviewsByDriver(l.ctx, driverID, page, pageSize)
	if err != nil {
		return nil, err
	}
	list := make([]types.ReviewItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, types.ReviewItem{
			OrderID:   int64(row.OrderID),
			UserID:    0, // 评价列表不暴露对端用户身份 ID，保护隐私
			DriverID:  int64(row.DriverID),
			Rating:    int32(row.Rating),
			Comment:   row.Comment,
			Tags:      row.Tags,
			CreatedAt: row.CreatedAt,
			Direction: "received",
		})
	}
	return &types.ListReceivedReviewsResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (l *ReviewLogic) reviewRepository() (svc.ReviewRepository, error) {
	if l.svcCtx == nil || l.svcCtx.ReviewRepository == nil {
		return nil, ErrReviewRepositoryNotConfigured
	}
	return l.svcCtx.ReviewRepository, nil
}

// minReviewOrders 是司机可接收乘客评价的完成订单门槛。
const minReviewOrders = 5

// GetReviewSummary 返回司机评价概览。服务平均分由 order_review 表真实聚合计算，
// 并返回完成单数、评价数、平均分与能否接收评价（完成订单 >= 5 单）。
func (l *ReviewLogic) GetReviewSummary(driverID int64) (*types.ReviewSummaryResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	repo, err := l.reviewRepository()
	if err != nil {
		return nil, err
	}
	reviewCount, avgRating, err := repo.CountAndAvgRatingByDriver(l.ctx, driverID)
	if err != nil {
		return nil, err
	}
	completed := l.completedOrderCount(l.ctx, driverID)
	canReceive := completed >= minReviewOrders

	return &types.ReviewSummaryResponse{
		CanReceiveReview:    canReceive,
		CompletedOrderCount: completed,
		ReviewCount:         reviewCount,
		AvgRating:           roundReviewRating(avgRating),
		ServiceScore:        roundReviewRating(avgRating),
	}, nil
}

// completedOrderCount 通过 ordersvc 统计司机已完成订单数（status=已完成）。
func (l *ReviewLogic) completedOrderCount(ctx context.Context, driverID int64) int64 {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return 0
	}
	resp, err := l.svcCtx.OrderClient.ListOrders(ctx, &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus_ORDER_STATUS_COMPLETED,
		Page:     1,
		PageSize: 1,
	})
	if err != nil || resp == nil {
		return 0
	}
	return resp.Total
}

// roundReviewRating 保留一位小数。
func roundReviewRating(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

// normalizeReviewPage 将分页参数规范为安全范围：页码最小 1，每页大小 1~100。
func normalizeReviewPage(page, pageSize int32) (int32, int32) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
