package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// 心跳/互踢相关的业务错误。
var (
	// errInvalidDriverID 表示司机 ID 非法。
	errInvalidDriverID = errors.New("司机ID不合法")
	// errInvalidDeviceID 表示设备标识缺失。
	errInvalidDeviceID = errors.New("设备标识缺失")
)

// HeartbeatLogic 司机心跳上报业务逻辑。
// 作用：刷新 Redis 在线状态 TTL 实现保活，并在心跳时机判定多端互踢（device_id 不匹配则踢出旧设备）。
type HeartbeatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewHeartbeatLogic 构造司机心跳上报逻辑处理器。
func NewHeartbeatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HeartbeatLogic {
	return &HeartbeatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Heartbeat 处理司机心跳：刷新在线状态保活，返回在线状态与是否被顶替（kicked）。
// 被顶替时返回 kicked=true，由客户端强制重新登录。
func (l *HeartbeatLogic) Heartbeat(in *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	// 校验司机 ID 与设备标识非空。
	if in.DriverId <= 0 {
		return nil, errInvalidDriverID
	}
	if in.DeviceId == "" {
		return nil, errInvalidDeviceID
	}
	// 调用 Redis 在线存储：刷新 TTL + 互踢判定。
	onlineStatus, kicked, err := l.svcCtx.OnlineStore.Heartbeat(l.ctx, in.DriverId, in.DeviceId, in.GetLongitude(), in.GetLatitude())
	if err != nil {
		// 在线存储异常：不阻断，返回当前状态，由调用方决定是否降级。
		return nil, err
	}
	return &proto.HeartbeatResponse{
		OnlineStatus: onlineStatus,
		Kicked:       kicked,
		ServerTime:   time.Now().Unix(),
	}, nil
}
