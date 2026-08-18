package logic

import (
	"context"

	"XiaoLong-Ridy/api/admin/internal/svc"
	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// StatisticsLogic 负责后台统计 HTTP 请求到 adminsvc 的参数转换。
type StatisticsLogic struct {
	ctx *svc.ServiceContext
}

// NewStatisticsLogic 创建统计逻辑对象。
func NewStatisticsLogic(ctx *svc.ServiceContext) *StatisticsLogic {
	return &StatisticsLogic{ctx: ctx}
}

// Overview 查询运营总览统计。
func (l *StatisticsLogic) Overview(ctx context.Context, req types.StatisticsRequest) (*adminclient.StatisticsOverviewResponse, error) {
	return l.ctx.AdminSvc.GetStatisticsOverview(ctx, statisticsRequestToPB(req))
}

// Orders 查询订单统计。
func (l *StatisticsLogic) Orders(ctx context.Context, req types.StatisticsRequest) (*adminclient.OrderStatisticsResponse, error) {
	return l.ctx.AdminSvc.GetOrderStatistics(ctx, statisticsRequestToPB(req))
}

// Coupons 查询优惠券统计。
func (l *StatisticsLogic) Coupons(ctx context.Context, req types.StatisticsRequest) (*adminclient.CouponStatisticsResponse, error) {
	return l.ctx.AdminSvc.GetCouponStatistics(ctx, statisticsRequestToPB(req))
}

// statisticsRequestToPB 将 HTTP 统计查询参数转换为 RPC 请求。
func statisticsRequestToPB(req types.StatisticsRequest) *adminclient.StatisticsRequest {
	return &adminclient.StatisticsRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		CityCode:  req.CityCode,
	}
}
