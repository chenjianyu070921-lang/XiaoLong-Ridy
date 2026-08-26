package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// errInvalidOnlineStatus 表示司机服务状态不是离线、在线或行程中。
var errInvalidOnlineStatus = errors.New("司机在线状态不合法")

// SetDriverServiceStatusLogic 封装司机服务状态同步逻辑。
// 该逻辑供司机端行程开始/结束链路调用，用于把司机状态同步到 Redis 在线存储、driver 表和 driver_location 表。
type SetDriverServiceStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSetDriverServiceStatusLogic 构造司机服务状态同步逻辑处理器。
func NewSetDriverServiceStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDriverServiceStatusLogic {
	return &SetDriverServiceStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetDriverServiceStatus 写入司机服务状态，供开始行程/结束行程等业务调用。
func (l *SetDriverServiceStatusLogic) SetDriverServiceStatus(in *proto.SetDriverServiceStatusRequest) (*proto.SetDriverServiceStatusResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if !validOnlineStatus(in.GetOnlineStatus()) {
		return nil, errInvalidOnlineStatus
	}
	if _, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.GetDriverId())); err != nil {
		return nil, err
	}

	statusValue := locationStatusFromOnline(in.GetOnlineStatus())
	if err := l.svcCtx.OnlineStore.SetStatus(l.ctx, in.GetDriverId(), in.GetOnlineStatus()); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.GetDriverId()), map[string]interface{}{"online_status": statusValue}); err != nil {
		return nil, err
	}
	if err := l.svcCtx.DriverRepository.UpdateLocationStatus(l.ctx, uint64(in.GetDriverId()), statusValue); err != nil {
		return nil, err
	}
	return &proto.SetDriverServiceStatusResponse{
		DriverId:     in.GetDriverId(),
		OnlineStatus: in.GetOnlineStatus(),
		UpdatedAt:    time.Now().Unix(),
	}, nil
}

// validOnlineStatus 校验服务状态是否为离线、在线或行程中。
func validOnlineStatus(onlineStatus int32) bool {
	switch onlineStatus {
	case onlinestore.Offline, onlinestore.Online, onlinestore.OnTrip:
		return true
	default:
		return false
	}
}

// locationStatusFromOnline 将在线存储状态映射为 driver_location.online_status。
func locationStatusFromOnline(onlineStatus int32) int8 {
	switch onlineStatus {
	case onlinestore.OnTrip:
		return model.LocationOnTrip
	case onlinestore.Online:
		return model.LocationOnline
	default:
		return model.LocationOffline
	}
}
