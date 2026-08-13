package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type ScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreLogic {
	return &ScoreLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *ScoreLogic) CreateScore(req *types.CreateScoreRequest) (*types.CreateScoreResponse, error) {
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if req.Score < 0 || req.Score > 100 {
		return nil, errors.New("服务分需在 0~100 之间")
	}
	if req.Level < 1 || req.Level > 5 {
		return nil, errors.New("司机等级需在 1~5 之间")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateScore(l.ctx, &driversproto.CreateScoreRequest{
		DriverId:           req.DriverID,
		Score:              req.Score,
		Level:              req.Level,
		MonthOrders:        req.MonthOrders,
		MonthCancelRate:    req.MonthCancelRate,
		MonthComplaintCount: req.MonthComplaintCount,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateScoreResponse{ID: resp.GetId(), DriverID: resp.GetDriverId()}, nil
}

func (l *ScoreLogic) UpdateScore(req *types.UpdateScoreRequest) (*types.UpdateScoreResponse, error) {
	if req.ID <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	if req.Score != nil && (*req.Score < 0 || *req.Score > 100) {
		return nil, errors.New("服务分需在 0~100 之间")
	}
	if req.Level != nil && (*req.Level < 1 || *req.Level > 5) {
		return nil, errors.New("司机等级需在 1~5 之间")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateScore(l.ctx, &driversproto.UpdateScoreRequest{
		Id:                 req.ID,
		DriverId:           req.DriverID,
		Score:              req.Score,
		Level:              req.Level,
		MonthOrders:        req.MonthOrders,
		MonthCancelRate:    req.MonthCancelRate,
		MonthComplaintCount: req.MonthComplaintCount,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateScoreResponse{ID: resp.GetId(), DriverID: resp.GetDriverId(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

func (l *ScoreLogic) GetScore(id int64) (*types.GetScoreResponse, error) {
	if id <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetScore(l.ctx, &driversproto.GetScoreRequest{Id: id})
	if err != nil {
		return nil, err
	}
	s := resp.GetScore()
	return &types.GetScoreResponse{Score: types.ScoreDetail{
		ID: s.GetId(), DriverID: s.GetDriverId(), Score: s.GetScore(), Level: s.GetLevel(),
		MonthOrders: s.GetMonthOrders(), MonthCancelRate: s.GetMonthCancelRate(),
		MonthComplaintCount: s.GetMonthComplaintCount(), UpdatedAt: s.GetUpdatedAt(),
	}}, nil
}

func (l *ScoreLogic) DeleteScore(id int64) (*types.DeleteResponse, error) {
	if id <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteScore(l.ctx, &driversproto.DeleteScoreRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

func (l *ScoreLogic) ListScores(req *types.ListScoresRequest) (*types.ListScoresResponse, error) {
	page, pageSize := clampPage(req.Page, req.PageSize)
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListScores(l.ctx, &driversproto.ListScoresRequest{
		DriverId: req.DriverID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.ScoreSummary, 0, len(resp.GetList()))
	for _, s := range resp.GetList() {
		list = append(list, types.ScoreSummary{
			ID: s.GetId(), DriverID: s.GetDriverId(), Score: s.GetScore(),
			Level: s.GetLevel(), MonthOrders: s.GetMonthOrders(),
		})
	}
	return &types.ListScoresResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

func (l *ScoreLogic) client() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
