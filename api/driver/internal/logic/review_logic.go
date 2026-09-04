package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// ErrReviewRepositoryNotConfigured 表示评价仓储未配置（MySQL 未初始化时降级）。
var ErrReviewRepositoryNotConfigured = errors.New("review repository not configured")

// ErrReviewAlreadyExists 表示同一订单已经评价过。
var ErrReviewAlreadyExists = errors.New("driver review already exists")

// ReviewLogic 封装司机端评价业务逻辑。
type ReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewReviewLogic 创建司机端评价逻辑处理器。
func NewReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLogic {
	return &ReviewLogic{ctx: ctx, svcCtx: svcCtx}
}

// ListReceivedReviews 返回当前司机收到的乘客评价（双向评价中的"接收乘客端评价"）。
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

// ListGivenReviews 返回当前司机给出的乘客评价。
func (l *ReviewLogic) ListGivenReviews(driverID int64, req *types.ListGivenReviewsRequest) (*types.ListGivenReviewsResponse, error) {
	if driverID <= 0 {
		return nil, ErrInvalidParam
	}
	page, pageSize := normalizeReviewPage(req.Page, req.PageSize)
	repo, err := l.reviewRepository()
	if err != nil {
		return nil, err
	}
	rows, total, err := repo.ListDriverReviewsByDriver(l.ctx, driverID, page, pageSize)
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
			CreatedAt: row.CreatedAt.Unix(),
			Direction: "given",
		})
	}
	return &types.ListGivenReviewsResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SubmitDriverReview 校验订单归属与完成状态后，写入司机对乘客的评价（双向评价中的"司机评乘客"）。
func (l *ReviewLogic) SubmitDriverReview(driverID int64, req *types.SubmitDriverReviewRequest) (*types.SubmitDriverReviewResponse, error) {
	if req == nil || req.OrderID <= 0 || req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidParam
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	repo, err := l.reviewRepository()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetDriverId() != driverID {
		return nil, ErrForbiddenDriverResource
	}
	if order.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_COMPLETED {
		return nil, ErrInvalidParam
	}
	review := &svc.DriverOrderReview{
		OrderID:   uint64(req.OrderID),
		UserID:    uint64(order.GetUserId()),
		DriverID:  uint64(driverID),
		Rating:    int8(req.Rating),
		Comment:   strings.TrimSpace(req.Comment),
		Tags:      strings.TrimSpace(req.Tags),
		CreatedAt: time.Now(),
	}
	if err := repo.CreateDriverReview(l.ctx, review); err != nil {
		if errors.Is(err, svc.ErrReviewAlreadyExists) {
			return nil, ErrReviewAlreadyExists
		}
		return nil, err
	}
	return &types.SubmitDriverReviewResponse{
		ReviewID:  int64(review.ID),
		OrderID:   int64(review.OrderID),
		DriverID:  int64(review.DriverID),
		Rating:    int32(review.Rating),
		CreatedAt: review.CreatedAt.Unix(),
	}, nil
}

func (l *ReviewLogic) reviewRepository() (svc.ReviewRepository, error) {
	if l.svcCtx == nil || l.svcCtx.ReviewRepository == nil {
		return nil, ErrReviewRepositoryNotConfigured
	}
	return l.svcCtx.ReviewRepository, nil
}

func (l *ReviewLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
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
