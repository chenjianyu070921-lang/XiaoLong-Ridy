package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetDriverAiScoreLogic 司机 AI 推荐得分业务逻辑。
// 基于已有司机服务分（driver_score）等运营指标，计算一个综合推荐分，
// 供司机端展示「自身 AI 推荐得分 / 影响因子」，并作为派单权重的参考输入。
type GetDriverAiScoreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetDriverAiScoreLogic 构造司机 AI 推荐得分逻辑处理器。
func NewGetDriverAiScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverAiScoreLogic {
	return &GetDriverAiScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 综合推荐分各维度的权重（权重之和=1，便于直观调整）。
const (
	weightServiceScore = 0.45 // 服务分（满分 100 归一化）
	weightCancelRate  = 0.25 // 取消率（越低越好）
	weightComplaint   = 0.15 // 投诉数（越少越好）
	weightMonthOrders = 0.15 // 当月完单数（越多越好，封顶 50 单）
)

// maxMonthOrders 当月完单数的归一化上限，超过视为满分贡献。
const maxMonthOrders = 50.0

// GetDriverAiScore 计算并返回指定司机的 AI 综合推荐得分与各项维度指标。
// 降级策略：若司机服务分记录不存在或查询异常，则返回 degraded=true，
// 回退到「距离优先」原始派单逻辑，保证打车链路不中断。
func (l *GetDriverAiScoreLogic) GetDriverAiScore(in *proto.GetDriverAiScoreRequest) (*proto.GetDriverAiScoreResponse, error) {
	if in.DriverId <= 0 {
		return nil, errors.New("司机ID不合法")
	}

	score, err := l.svcCtx.DriverRepository.GetDriverScore(l.ctx, uint64(in.DriverId))
	if err != nil {
		// 查询异常：降级，不阻断业务。
		return degradedResponse(in.DriverId, "服务分查询失败，已回退距离优先策略"), nil
	}
	if score == nil {
		// 无服务分记录（新司机/未统计）：降级，给出默认提示。
		return degradedResponse(in.DriverId, "暂无服务分记录，已回退距离优先策略"), nil
	}

	// 1) 服务分：满分 100 直接归一化，超出按 100 计。
	serviceNorm := clamp01(score.Score / 100.0)
	// 2) 取消率：越低越好，0% 得满分，100% 得 0 分。
	cancelNorm := clamp01(1.0 - score.MonthCancelRate/100.0)
	// 3) 投诉数：0 投诉满分，每 1 次投诉扣 20 分（封底 0）。
	complaintNorm := clamp01(1.0 - float64(score.MonthComplaintCount)*0.2)
	// 4) 当月完单数：越多越好，封顶 maxMonthOrders 得满分。
	orderNorm := clamp01(float64(score.MonthOrders) / maxMonthOrders)

	// 加权求和得到 0~100 的综合推荐分。
	aiScore := (serviceNorm*weightServiceScore +
		cancelNorm*weightCancelRate +
		complaintNorm*weightComplaint +
		orderNorm*weightMonthOrders) * 100.0

	factors := []*proto.AiScoreFactor{
		{
			Key:     "service_score",
			Label:   "服务分",
			Value:   score.Score,
			Impact:  impactOf(serviceNorm, "提升服务分可提升推荐优先级", "服务分偏低，建议规范服务提升得分"),
			Hint:    "保持高服务分（好评、准时）可显著提升推荐优先级",
		},
		{
			Key:     "cancel_rate",
			Label:   "当月取消率",
			Value:   score.MonthCancelRate,
			Impact:  impactOf(cancelNorm, "低取消率有利于推荐优先级", "取消率偏高，降低推荐优先级"),
			Hint:    "降低订单取消率可提升推荐优先级",
		},
		{
			Key:     "complaint",
			Label:   "当月投诉数",
			Value:   float64(score.MonthComplaintCount),
			Impact:  impactOf(complaintNorm, "无投诉有利于推荐优先级", "存在投诉，降低推荐优先级"),
			Hint:    "减少服务投诉可提升推荐优先级",
		},
		{
			Key:     "month_orders",
			Label:   "当月完单数",
			Value:   float64(score.MonthOrders),
			Impact:  impactOf(orderNorm, "完单数多有利于推荐优先级", "完单数偏少，推荐优先级较低"),
			Hint:    "提高完单量可提升推荐优先级",
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

// impactOf 根据维度归一化值给出影响说明：>=0.6 视为正面，<=0.3 视为负面，中间中性。
func impactOf(norm float64, positive, negative string) string {
	switch {
	case norm >= 0.6:
		return "positive"
	case norm <= 0.3:
		return "negative"
	default:
		return "neutral"
	}
}

// degradedResponse 构造降级响应，业务链路不中断。
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

// clamp01 将数值收敛到 [0,1]。
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// round1 保留一位小数。
func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
