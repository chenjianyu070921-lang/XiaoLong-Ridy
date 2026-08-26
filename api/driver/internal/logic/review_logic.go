package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

type ReviewLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewLogic {
	return &ReviewLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ReviewLogic) ListPassengerReviews(driverID int64, req *types.ListPassengerReviewsRequest) (*types.ListPassengerReviewsResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	if l.svcCtx == nil || l.svcCtx.ReviewRepository == nil {
		return nil, ErrReviewStorageNotConfigured
	}
	page, pageSize := clampPage(req.Page, req.PageSize)
	records, total, err := l.svcCtx.ReviewRepository.ListByDriver(l.ctx, driverID, page, pageSize)
	if err != nil {
		return nil, err
	}
	resp := &types.ListPassengerReviewsResponse{
		List:     make([]types.PassengerReview, 0, len(records)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	for _, record := range records {
		resp.List = append(resp.List, types.PassengerReview{
			OrderID:   record.OrderID,
			Rating:    record.Rating,
			Comment:   record.Comment,
			CreatedAt: record.CreatedAt.Unix(),
		})
	}
	return resp, nil
}
