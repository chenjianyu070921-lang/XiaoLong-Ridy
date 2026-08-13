package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListWithdrawsLogic 处理分页查询提现列表请求的逻辑结构体。
type ListWithdrawsLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewListWithdrawsLogic 构造 ListWithdrawsLogic 实例。
func NewListWithdrawsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListWithdrawsLogic {
	return &ListWithdrawsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListWithdraws 分页查询提现列表，支持按司机、状态过滤。
func (l *ListWithdrawsLogic) ListWithdraws(in *proto.ListWithdrawsRequest) (*proto.ListWithdrawsResponse, error) {
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
	query := l.svcCtx.DB.Model(&model.DriverWithdraw{})
	if in.DriverId != 0 {
		query = query.Where("driver_id = ?", in.DriverId) // 按司机过滤
	}
	if in.Status != 0 {
		query = query.Where("status = ?", in.Status) // 按状态过滤
	}

	// 统计符合条件的总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询，按 ID 倒序返回
	var withdraws []model.DriverWithdraw
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&withdraws).Error; err != nil {
		return nil, err
	}

	// 转换为精简的提现摘要列表
	list := make([]*proto.WithdrawSummary, 0, len(withdraws))
	for _, w := range withdraws {
		list = append(list, &proto.WithdrawSummary{
			Id:          int64(w.Id),              // 提现 ID
			DriverId:    int64(w.DriverId),        // 所属司机 ID
			WithdrawNo:  w.WithdrawNo,             // 提现单号
			Amount:      w.Amount,                 // 提现金额
			Status:      int32(w.Status),          // 状态
			AppliedAt:   unixOrZero(w.AppliedAt),  // 申请时间
		})
	}

	// 返回列表、总数与分页信息
	return &proto.ListWithdrawsResponse{
		List:     list,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
