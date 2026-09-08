package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

var (
	// ErrReviewRepositoryNotConfigured 表示评价仓储未配置。
	ErrReviewRepositoryNotConfigured = errors.New("review repository not configured")
	// ErrReviewAlreadyExists 表示同一订单已经评价过。
	ErrReviewAlreadyExists = errors.New("review already exists")
	// ErrReviewNotQualified 表示司机完成订单数不足，暂不支持乘客评价。
	ErrReviewNotQualified = errors.New("driver has not enough completed orders to receive review")
)

// minReviewCompletedOrders 是司机可接收乘客评价的完成订单门槛。
const minReviewCompletedOrders = 5

// ReviewLogic 封装乘客评价订单的业务流程。
type ReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewReviewLogic 创建评价业务逻辑对象。
func NewReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *ReviewLogic {
	return &ReviewLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// SubmitReview 校验订单归属和完成状态后写入乘客评价。
func (l *ReviewLogic) SubmitReview(req *types.SubmitReviewRequest) (*types.SubmitReviewResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.OrderID <= 0 || req.Rating < 1 || req.Rating > 5 {
		return nil, ErrInvalidRequest
	}
	orderClient, err := l.orderClient()
	if err != nil {
		return nil, err
	}
	reviews, err := l.reviewRepository()
	if err != nil {
		return nil, err
	}
	order, err := orderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: req.OrderID})
	if err != nil {
		return nil, err
	}
	if order.GetUserId() != int64(userID) {
		return nil, ErrForbidden
	}
	if order.GetStatus() != orderproto.OrderStatus_ORDER_STATUS_COMPLETED || order.GetDriverId() <= 0 {
		return nil, ErrInvalidRequest
	}
	// 门槛限制：司机完成订单未满 5 单时，乘客不能对该司机评价。
	completedOrders, err := l.driverCompletedOrders(order.GetDriverId())
	if err != nil {
		return nil, err
	}
	if completedOrders < minReviewCompletedOrders {
		return nil, ErrReviewNotQualified
	}
	review := &svc.OrderReview{
		OrderID:   uint64(req.OrderID),
		UserID:    userID,
		DriverID:  uint64(order.GetDriverId()),
		Rating:    int8(req.Rating),
		Comment:   strings.TrimSpace(req.Comment),
		Tags:      strings.TrimSpace(req.Tags),
		CreatedAt: time.Now(),
	}
	if err := reviews.Create(l.ctx, review); err != nil {
		if errors.Is(err, svc.ErrReviewAlreadyExists) {
			return nil, ErrReviewAlreadyExists
		}
		return nil, err
	}
	return &types.SubmitReviewResponse{
		ReviewID:  int64(review.ID),
		OrderID:   int64(review.OrderID),
		DriverID:  int64(review.DriverID),
		Rating:    int32(review.Rating),
		CreatedAt: review.CreatedAt.Unix(),
	}, nil
}

// driverCompletedOrders 返回司机已完成订单总数（status=已完成）。
func (l *ReviewLogic) driverCompletedOrders(driverID int64) (int64, error) {
	orderClient, err := l.orderClient()
	if err != nil {
		return 0, err
	}
	resp, err := orderClient.ListOrders(l.ctx, &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus_ORDER_STATUS_COMPLETED,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, nil
	}
	return resp.Total, nil
}

// reviewRepository 获取评价仓储依赖。
func (l *ReviewLogic) reviewRepository() (svc.ReviewRepository, error) {
	if l.svcCtx == nil || l.svcCtx.Reviews == nil {
		return nil, ErrReviewRepositoryNotConfigured
	}
	return l.svcCtx.Reviews, nil
}

// orderClient 获取订单服务客户端，用于校验评价所属订单和订单完成状态。
func (l *ReviewLogic) orderClient() (svc.OrderClient, error) {
	if l.svcCtx == nil || l.svcCtx.OrderClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.OrderClient, nil
}
