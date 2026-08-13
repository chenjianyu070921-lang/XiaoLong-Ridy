// Package logic 实现 driver API 的业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"   // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"  // API 层使用的请求/响应类型
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// ScoreLogic 服务分业务逻辑处理器，持有上下文与下游客户端。
type ScoreLogic struct {
	ctx    context.Context    // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewScoreLogic 构造服务分逻辑处理器实例。
func NewScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ScoreLogic {
	// 注入上下文与服务上下文。
	return &ScoreLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateScore 创建服务分，校验司机归属、分数与等级范围。
func (l *ScoreLogic) CreateScore(req *types.CreateScoreRequest) (*types.CreateScoreResponse, error) {
	// 校验所属司机 ID 合法性。
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 校验服务分在 0~100 之间。
	if req.Score < 0 || req.Score > 100 {
		return nil, errors.New("服务分需在 0~100 之间")
	}
	// 校验司机等级在 1~5 之间。
	if req.Level < 1 || req.Level > 5 {
		return nil, errors.New("司机等级需在 1~5 之间")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游创建服务分接口。
	resp, err := client.CreateScore(l.ctx, &driversproto.CreateScoreRequest{
		DriverId:            req.DriverID,            // 司机 ID
		Score:               req.Score,               // 服务分
		Level:               req.Level,               // 等级
		MonthOrders:         req.MonthOrders,         // 当月完单数
		MonthCancelRate:     req.MonthCancelRate,     // 当月取消率
		MonthComplaintCount: req.MonthComplaintCount, // 当月投诉数
	})
	if err != nil {
		return nil, err
	}
	// 返回创建结果（记录 ID + 司机 ID）。
	return &types.CreateScoreResponse{ID: resp.GetId(), DriverID: resp.GetDriverId()}, nil
}

// UpdateScore 更新服务分，校验 ID 与可选分数/等级范围。
func (l *ScoreLogic) UpdateScore(req *types.UpdateScoreRequest) (*types.UpdateScoreResponse, error) {
	// 校验记录 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	// 若传入分数，校验其范围 0~100。
	if req.Score != nil && (*req.Score < 0 || *req.Score > 100) {
		return nil, errors.New("服务分需在 0~100 之间")
	}
	// 若传入等级，校验其范围 1~5。
	if req.Level != nil && (*req.Level < 1 || *req.Level > 5) {
		return nil, errors.New("司机等级需在 1~5 之间")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口，可选字段直接透传指针。
	resp, err := client.UpdateScore(l.ctx, &driversproto.UpdateScoreRequest{
		Id:                 req.ID,                 // 记录 ID
		DriverId:           req.DriverID,           // 可选司机 ID
		Score:              req.Score,              // 可选服务分
		Level:              req.Level,              // 可选等级
		MonthOrders:        req.MonthOrders,        // 可选当月完单数
		MonthCancelRate:    req.MonthCancelRate,    // 可选当月取消率
		MonthComplaintCount: req.MonthComplaintCount, // 可选当月投诉数
	})
	if err != nil {
		return nil, err
	}
	// 返回更新结果。
	return &types.UpdateScoreResponse{ID: resp.GetId(), DriverID: resp.GetDriverId(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetScore 查询服务分详情。
func (l *ScoreLogic) GetScore(id int64) (*types.GetScoreResponse, error) {
	// 校验记录 ID 合法性。
	if id <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetScore(l.ctx, &driversproto.GetScoreRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的服务分实体。
	s := resp.GetScore()
	// 映射为 API 的服务分详情结构并返回。
	return &types.GetScoreResponse{Score: types.ScoreDetail{
		ID:                 s.GetId(),
		DriverID:           s.GetDriverId(),
		Score:              s.GetScore(),
		Level:              s.GetLevel(),
		MonthOrders:        s.GetMonthOrders(),
		MonthCancelRate:    s.GetMonthCancelRate(),
		MonthComplaintCount: s.GetMonthComplaintCount(),
		UpdatedAt:          s.GetUpdatedAt(),
	}}, nil
}

// DeleteScore 删除服务分记录。
func (l *ScoreLogic) DeleteScore(id int64) (*types.DeleteResponse, error) {
	// 校验记录 ID 合法性。
	if id <= 0 {
		return nil, errors.New("服务分记录ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游删除接口。
	resp, err := client.DeleteScore(l.ctx, &driversproto.DeleteScoreRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 返回删除结果。
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// ListScores 分页查询服务分列表。
func (l *ScoreLogic) ListScores(req *types.ListScoresRequest) (*types.ListScoresResponse, error) {
	// 收敛分页参数到合法范围。
	page, pageSize := clampPage(req.Page, req.PageSize)
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游列表接口。
	resp, err := client.ListScores(l.ctx, &driversproto.ListScoresRequest{
		DriverId: req.DriverID, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	// 预分配切片。
	list := make([]types.ScoreSummary, 0, len(resp.GetList()))
	// 遍历并映射为 API 摘要结构。
	for _, s := range resp.GetList() {
		list = append(list, types.ScoreSummary{
			ID:          s.GetId(),
			DriverID:    s.GetDriverId(),
			Score:       s.GetScore(),
			Level:       s.GetLevel(),
			MonthOrders: s.GetMonthOrders(),
		})
	}
	// 组装分页响应返回。
	return &types.ListScoresResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// client 从服务上下文中安全取出 driversvc 客户端。
func (l *ScoreLogic) client() (svc.DriverClient, error) {
	// 防御性校验客户端是否可用。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
