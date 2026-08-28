package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ListDriversLogic 司机列表查询逻辑。
type ListDriversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListDriversLogic 构造司机列表查询逻辑处理器。
func NewListDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriversLogic {
	return &ListDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListDrivers 分页查询司机列表，支持状态与关键字过滤。
func (l *ListDriversLogic) ListDrivers(in *proto.ListDriversRequest) (*proto.ListDriversResponse, error) {
	if in == nil {
		return nil, errors.New("请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	// 组装过滤条件；状态为空时不过滤。
	filter := repository.DriverListFilter{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Keyword:  in.GetKeyword(),
	}
	if in.Status != nil {
		status := int8(in.GetStatus())
		filter.Status = &status
	}

	drivers, total, err := l.svcCtx.DriverRepository.List(l.ctx, filter)
	if err != nil {
		return nil, err
	}

	list := make([]*proto.Driver, 0, len(drivers))
	for _, d := range drivers {
		list = append(list, &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			PasswordHash:    d.PasswordHash,
			RealName:        d.RealName,
			IdCardNo:        d.IdCardNo,
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			OnlineStatus:    int32(d.OnlineStatus),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		})
	}
	return &proto.ListDriversResponse{
		Drivers: list,
		Total:   total,
	}, nil
}
