package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/clause"
)

// CreateScoreLogic 处理创建/初始化服务分请求的逻辑结构体。
type CreateScoreLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewCreateScoreLogic 构造 CreateScoreLogic 实例。
func NewCreateScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateScoreLogic {
	return &CreateScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateScore 初始化司机服务分，采用 upsert：同一司机已存在则覆盖运营指标。
// 默认服务分 100、等级 1，可由请求覆盖。
func (l *CreateScoreLogic) CreateScore(in *proto.CreateScoreRequest) (*proto.CreateScoreResponse, error) {
	score := in.Score
	if score == 0 {
		score = 100 // 默认服务分 100
	}
	level := in.Level
	if level == 0 {
		level = 1 // 默认等级 1
	}

	s := &model.DriverScore{
		DriverId:           uint64(in.DriverId),           // 所属司机 ID
		Score:              score,                         // 服务分
		Level:              int8(level),                   // 司机等级
		MonthOrders:        int(in.MonthOrders),           // 当月完单数
		MonthCancelRate:    in.MonthCancelRate,            // 当月取消率
		MonthComplaintCount: int(in.MonthComplaintCount),  // 当月投诉数
	}
	// upsert：driver_id 唯一，存在则更新运营指标
	if err := l.svcCtx.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "driver_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "level", "month_orders", "month_cancel_rate", "month_complaint_count"}),
	}).Create(s).Error; err != nil {
		return nil, err
	}
	// 返回新建/更新记录 ID 与司机 ID
	return &proto.CreateScoreResponse{
		Id:        int64(s.Id),
		DriverId:  in.DriverId,
	}, nil
}
