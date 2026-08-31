package logic

import (
	"context"
	"errors"
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DriverOffline int8 = 0
	DriverOnline  int8 = 1
)

type SetDriverOnlineLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDriverOnlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDriverOnlineLogic {
	return &SetDriverOnlineLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetDriverOnlineLogic) SetDriverOnline(in *proto.SetDriverOnlineRequest) (*proto.SetDriverOnlineResponse, error) {
	if in == nil || in.GetDriverId() <= 0 {
		return nil, errInvalidDriverID
	}
	if in.GetDeviceId() == "" {
		return nil, errInvalidDeviceID
	}
	if !validLongitudeLatitude(in.GetLongitude(), in.GetLatitude()) {
		return nil, errInvalidLongitudeLatitude
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil || l.svcCtx.OnlineStore == nil {
		return nil, errors.New("driver dependencies not ready")
	}
	driver, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.DriverId))
	if err != nil {
		return nil, err
	}
	if driver == nil {
		return nil, status.Error(codes.PermissionDenied, "driver account is not active")
	}
	if driver.Status != int8(proto.DriverStatus_DRIVER_STATUS_NORMAL) {
		return nil, status.Error(codes.PermissionDenied, "driver account is not active")
	}
	if l.svcCtx.CertificationRepository == nil {
		return nil, errors.New("driver dependencies not ready")
	}
	cert, err := l.svcCtx.CertificationRepository.GetByDriverID(l.ctx, uint64(in.DriverId))
	if err != nil {
		if errors.Is(err, repository.ErrCertificationNotFound) {
			return nil, status.Error(codes.PermissionDenied, "driver certification not approved")
		}
		return nil, err
	}
	if cert == nil || cert.AuditStatus != AuditStatusPassed {
		return nil, status.Error(codes.PermissionDenied, "driver certification not approved")
	}
	if err := l.svcCtx.OnlineStore.SetOnline(l.ctx, in.GetDriverId(), in.GetDeviceId(), in.GetLongitude(), in.GetLatitude()); err != nil {
		return nil, err
	}
	reportTime := time.Now()
	if err := l.svcCtx.DriverRepository.UpsertLocation(l.ctx, &model.DriverLocation{
		DriverID:     uint64(in.GetDriverId()),
		Longitude:    in.GetLongitude(),
		Latitude:     in.GetLatitude(),
		OnlineStatus: DriverOnline,
		ReportTime:   reportTime,
	}); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{"online_status": DriverOnline}
	if err := l.svcCtx.DriverRepository.Update(l.ctx, uint64(in.DriverId), updates); err != nil {
		return nil, err
	}
	// 派单侧同步失败仅告警，不阻断主流程（onlinestore 和 DB 已更新，避免司机端收到"上线失败"但实际已上线）
	// 先清理上次残留的 busy 标记（异常订单/服务重启可能遗留），否则司机即使上线也会被派单引擎过滤。
	// 注意：busy 清理只能在上线时做，心跳/位置上报不能清，否则服务中司机会被重复派单。
	if err := clearDispatchDriverBusy(l.ctx, l.svcCtx, in.GetDriverId()); err != nil {
		l.Errorf("clear dispatch driver busy failed (driver already online in DB/Redis): %v", err)
	}
	if err := syncDispatchDriverOnline(l.ctx, l.svcCtx, in.GetDriverId(), in.GetLongitude(), in.GetLatitude()); err != nil {
		l.Errorf("sync dispatch online failed (driver already online in DB/Redis): %v", err)
	}
	return &proto.SetDriverOnlineResponse{
		DriverId:     in.DriverId,
		OnlineStatus: int32(DriverOnline),
	}, nil
}
