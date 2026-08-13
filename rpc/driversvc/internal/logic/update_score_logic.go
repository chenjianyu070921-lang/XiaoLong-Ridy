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

// UpdateScoreLogic 处理更新服务分请求的逻辑结构体。
type UpdateScoreLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewUpdateScoreLogic 构造 UpdateScoreLogic 实例。
func NewUpdateScoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateScoreLogic {
	return &UpdateScoreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateScore 更新服务分，仅修改请求中显式传入的字段。
func (l *UpdateScoreLogic) UpdateScore(in *proto.UpdateScoreRequest) (*proto.UpdateScoreResponse, error) {
	// 先按 ID 查询记录是否存在
	var s model.DriverScore
	err := l.svcCtx.DB.First(&s, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("score not found") // 记录不存在
	}
	if err != nil {
		return nil, err
	}

	// 仅更新显式提供的字段（optional 字段为指针，nil 表示不更新）
	updates := map[string]interface{}{}
	if in.DriverId != nil {
		updates["driver_id"] = in.GetDriverId() // 所属司机 ID
	}
	if in.Score != nil {
		updates["score"] = in.GetScore() // 服务分
	}
	if in.Level != nil {
		updates["level"] = int8(in.GetLevel()) // 司机等级
	}
	if in.MonthOrders != nil {
		updates["month_orders"] = int(in.GetMonthOrders()) // 当月完单数
	}
	if in.MonthCancelRate != nil {
		updates["month_cancel_rate"] = in.GetMonthCancelRate() // 当月取消率
	}
	if in.MonthComplaintCount != nil {
		updates["month_complaint_count"] = int(in.GetMonthComplaintCount()) // 当月投诉数
	}

	// 执行更新
	if err := l.svcCtx.DB.Model(&s).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新读取更新后的记录
	if err := l.svcCtx.DB.First(&s, in.Id).Error; err != nil {
		return nil, err
	}
	// 返回更新后的 ID、司机 ID 与更新时间
	return &proto.UpdateScoreResponse{
		Id:        int64(s.Id),
		DriverId:  int64(s.DriverId),
		UpdatedAt: s.UpdatedAt.Unix(),
	}, nil
}
