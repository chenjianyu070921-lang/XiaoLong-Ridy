package logic

import (
	"context"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

type TrajectoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTrajectoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TrajectoryLogic {
	return &TrajectoryLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *TrajectoryLogic) GetTripTrajectory(driverID int64, req *types.GetTripTrajectoryRequest) (*types.GetTripTrajectoryResponse, error) {
	if driverID <= 0 || req == nil || req.OrderID <= 0 {
		return nil, ErrInvalidParam
	}
	if l.svcCtx == nil || l.svcCtx.TrajectoryRepository == nil {
		return nil, ErrTrajectoryStorageNotConfigured
	}
	records, err := l.svcCtx.TrajectoryRepository.ListByOrder(l.ctx, driverID, req.OrderID)
	if err != nil {
		return nil, err
	}
	resp := &types.GetTripTrajectoryResponse{
		OrderID: req.OrderID,
		Points:  make([]types.TrajectoryPoint, 0, len(records)),
	}
	for _, record := range records {
		resp.Points = append(resp.Points, types.TrajectoryPoint{
			Longitude: record.Longitude,
			Latitude:  record.Latitude,
			SpeedKmh:  record.SpeedKmh,
			Heading:   record.Heading,
			CreatedAt: record.RecordedAt.Unix(),
		})
	}
	return resp, nil
}
