package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListDriversLogic 处理分页查询司机列表请求的逻辑结构体。
type ListDriversLogic struct {
	ctx    context.Context      // ctx：请求上下文
	svcCtx *svc.ServiceContext  // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewListDriversLogic 构造 ListDriversLogic 实例。
func NewListDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriversLogic {
	return &ListDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDrivers 分页查询司机列表，支持按状态、手机号关键字过滤。
func (l *ListDriversLogic) ListDrivers(in *proto.ListDriversRequest) (*proto.ListDriversResponse, error) {
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
	query := l.svcCtx.DB.Model(&model.Driver{})
	if in.Status != proto.DriverStatus_DRIVER_STATUS_UNSPECIFIED {
		query = query.Where("status = ?", int8(in.Status)) // 按账号状态过滤
	}
	if in.PhoneKeyword != "" {
		query = query.Where("phone LIKE ?", "%"+in.PhoneKeyword+"%") // 按手机号模糊匹配
	}

	// 统计符合条件的总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询，按 ID 倒序返回
	var drivers []model.Driver
	if err := query.Order("id DESC").
		Offset((page - 1) * pageSize). // 跳过的记录数
		Limit(pageSize).               // 本页记录数
		Find(&drivers).Error; err != nil {
		return nil, err
	}

	// 转换为精简的司机摘要列表
	list := make([]*proto.DriverSummary, 0, len(drivers))
	for _, d := range drivers {
		list = append(list, &proto.DriverSummary{
			Id:              int64(d.Id),                    // 司机 ID
			Phone:           d.Phone,                       // 手机号
			RealName:        d.RealName,                    // 真实姓名
			DriverLicenseNo: d.DriverLicenseNo,             // 驾驶证号
			Status:          proto.DriverStatus(d.Status), // 账号状态
			CreatedAt:       d.CreatedAt.Unix(),            // 创建时间（Unix 秒）
		})
	}

	// 返回列表、总数与分页信息
	return &proto.ListDriversResponse{
		List:     list,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}
