package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshDriverScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshDriverScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshDriverScoreLogic {
	return &RefreshDriverScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重新计算并写回司机评分指标（喂数），返回刷新后的评分，供联调与定时任务触发。
// 这是「AI 评分无模块内喂数」的修复入口：RefreshDriverScoreMetrics 此前没有任何调用方，
// 导致 driver_score 长期无数据，GetDriverAiScore 只能恒降级为距离优先。
func (l *RefreshDriverScoreLogic) RefreshDriverScore(in *proto.RefreshDriverScoreRequest) (*proto.GetDriverAiScoreResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errors.New("invalid driver id")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}

	// 按自然月重算，与 RefreshDriverScoreMetrics 的统计口径（本月完成单/取消率/投诉数）保持一致。
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if _, err := l.svcCtx.DriverRepository.RefreshDriverScoreMetrics(l.ctx, uint64(in.DriverId), monthStart, now); err != nil {
		return degradedResponse(in.DriverId, "AI score refresh failed; falling back to distance priority"), nil
	}

	// 复用查询口径组装响应，保证刷新结果与 GetDriverAiScore 完全一致。
	return NewGetDriverAiScoreLogic(l.ctx, l.svcCtx).GetDriverAiScore(&proto.GetDriverAiScoreRequest{DriverId: in.DriverId})
}
