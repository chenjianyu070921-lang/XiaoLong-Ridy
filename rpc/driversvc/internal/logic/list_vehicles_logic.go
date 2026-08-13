package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListVehiclesLogic 处理分页查询车辆列表请求的逻辑结构体。
type ListVehiclesLogic struct {
	ctx    context.Context     // ctx：请求上下文
	svcCtx *svc.ServiceContext // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewListVehiclesLogic 构造 ListVehiclesLogic 实例。
func NewListVehiclesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVehiclesLogic {
	return &ListVehiclesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListVehicles 分页查询车辆列表，支持按司机、状态过滤。
func (l *ListVehiclesLogic) ListVehicles(in *proto.ListVehiclesRequest) (*proto.ListVehiclesResponse, error) {
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
	query := l.svcCtx.DB.Model(&model.DriverVehicle{})
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
	var vehicles []model.DriverVehicle
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&vehicles).Error; err != nil {
		return nil, err
	}

	// 转换为精简的车辆摘要列表
	list := make([]*proto.VehicleSummary, 0, len(vehicles))
	for _, v := range vehicles {
		list = append(list, &proto.VehicleSummary{
			Id:          int64(v.Id),        // 车辆 ID
			DriverId:    int64(v.DriverId),  // 所属司机 ID
			PlateNo:     v.PlateNo,          // 车牌号
			Brand:       v.Brand,            // 品牌
			VehicleType: int32(v.VehicleType), // 车辆类型
			Status:      int32(v.Status),    // 状态
		})
	}

	// 返回列表、总数与分页信息
	return &proto.ListVehiclesResponse{
		List:     list,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
