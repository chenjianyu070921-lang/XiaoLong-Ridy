package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// GetScoreLogic 处理查询服务分详情请求的逻辑结构体。
type GetScoreLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewGetScoreLogic 构造 GetScoreLogic 实例。
func NewGetScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScoreLogic {
	return &GetScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetScore 根据记录 ID 查询服务分完整信息。
func (l *GetScoreLogic) GetScore(in *proto.GetScoreRequest) (*proto.GetScoreResponse, error) {
	// 按 ID 查询记录
	var s model.DriverScore
	err := l.svcCtx.DB.First(&s, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("score not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}
	// 组装并返回服务分详情
	return &proto.GetScoreResponse{
		Score: &proto.Score{
			Id:                  int64(s.Id),                 // 记录 ID
			DriverId:            int64(s.DriverId),           // 所属司机 ID
			Score:               s.Score,                     // 服务分
			Level:               int32(s.Level),              // 司机等级
			MonthOrders:         int32(s.MonthOrders),         // 当月完单数
			MonthCancelRate:     s.MonthCancelRate,            // 当月取消率
			MonthComplaintCount: int32(s.MonthComplaintCount), // 当月投诉数
			UpdatedAt:           s.UpdatedAt.Unix(),           // 更新时间
		},
	}, nil
}
