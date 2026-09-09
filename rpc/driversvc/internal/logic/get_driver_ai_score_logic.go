package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDriverAiScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverAiScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverAiScoreLogic {
	return &GetDriverAiScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

const (
	weightServiceScore = 0.45
	weightCancelRate   = 0.25
	weightComplaint    = 0.15
	weightMonthOrders  = 0.15
	maxMonthOrders     = 50.0
)

func (l *GetDriverAiScoreLogic) GetDriverAiScore(in *proto.GetDriverAiScoreRequest) (*proto.GetDriverAiScoreResponse, error) {
	if in == nil || in.DriverId <= 0 {
		return nil, errors.New("invalid driver id")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}

	score, err := l.svcCtx.DriverRepository.GetDriverScore(l.ctx, uint64(in.DriverId))
	if err != nil {
		return degradedResponse(in.DriverId, "AI score lookup failed; falling back to distance priority"), nil
	}
	if score == nil {
		return degradedResponse(in.DriverId, "no AI score metrics available; falling back to distance priority"), nil
	}

	serviceNorm := clamp01(score.Score / 100.0)
	cancelNorm := clamp01(1.0 - score.MonthCancelRate/100.0)
	complaintNorm := clamp01(1.0 - float64(score.MonthComplaintCount)*0.2)
	orderNorm := clamp01(float64(score.MonthOrders) / maxMonthOrders)

	aiScore := (serviceNorm*weightServiceScore +
		cancelNorm*weightCancelRate +
		complaintNorm*weightComplaint +
		orderNorm*weightMonthOrders) * 100.0

	factors := []*proto.AiScoreFactor{
		{
			Key:    "service_score",
			Label:  "service_score",
			Value:  score.Score,
			Impact: impactOf(serviceNorm, "positive", "negative"),
			Hint:   "higher service score improves recommendation priority",
		},
		{
			Key:    "cancel_rate",
			Label:  "cancel_rate",
			Value:  score.MonthCancelRate,
			Impact: impactOf(cancelNorm, "positive", "negative"),
			Hint:   "lower cancel rate improves recommendation priority",
		},
		{
			Key:    "complaint",
			Label:  "complaint",
			Value:  float64(score.MonthComplaintCount),
			Impact: impactOf(complaintNorm, "positive", "negative"),
			Hint:   "fewer complaints improve recommendation priority",
		},
		{
			Key:    "month_orders",
			Label:  "month_orders",
			Value:  float64(score.MonthOrders),
			Impact: impactOf(orderNorm, "positive", "negative"),
			Hint:   "more completed orders improve recommendation priority",
		},
	}

	return &proto.GetDriverAiScoreResponse{
		DriverId: in.DriverId,
		AiScore:  round1(aiScore),
		Level:    int32(score.Level),
		Factors:  factors,
		Degraded: false,
	}, nil
}

func impactOf(norm float64, positive, negative string) string {
	switch {
	case norm >= 0.6:
		return positive
	case norm <= 0.3:
		return negative
	default:
		return "neutral"
	}
}

func degradedResponse(driverID int64, reason string) *proto.GetDriverAiScoreResponse {
	return &proto.GetDriverAiScoreResponse{
		DriverId:      driverID,
		AiScore:       0,
		Level:         0,
		Factors:       nil,
		Degraded:      true,
		DegradeReason: reason,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
