package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListScoresLogic 处理分页查询服务分列表请求的逻辑结构体。
type ListScoresLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewListScoresLogic 构造 ListScoresLogic 实例。
func NewListScoresLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListScoresLogic {
	return &ListScoresLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListScores 分页查询服务分列表，支持按司机过滤。
func (l *ListScoresLogic) ListScores(in *proto.ListScoresRequest) (*proto.ListScoresResponse, error) {
	// 解析分页参数，并做默认值与上限保护
	page := int(in.Page)
	if page <= 0 {
		page = 1 // 默认第 1 页
	}
	pageSize := int(in.PageSize)
	if pageSize <= 0 {
		pageSize = 20 // 默认每页 20 条
	}
	if pageSize > 100 {
		pageSize = 100 // 每页最多 100 条
	}

	// 构建查询条件
	query := l.svcCtx.DB.Model(&model.DriverScore{})
	if in.DriverId != 0 {
		query = query.Where("driver_id = ?", in.DriverId) // 按司机过滤
	}

	// 统计符合条件的总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询，按 ID 倒序返回
	var scores []model.DriverScore
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&scores).Error; err != nil {
		return nil, err
	}

	// 转换为精简的服务分摘要列表
	list := make([]*proto.ScoreSummary, 0, len(scores))
	for _, s := range scores {
		list = append(list, &proto.ScoreSummary{
			Id:            int64(s.Id),             // 记录 ID
			DriverId:      int64(s.DriverId),       // 所属司机 ID
			Score:         s.Score,                 // 服务分
			Level:         int32(s.Level),          // 司机等级
			MonthOrders:   int32(s.MonthOrders),     // 当月完单数
		})
	}

	// 返回列表、总数与分页信息
	return &proto.ListScoresResponse{
		List:     list,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
