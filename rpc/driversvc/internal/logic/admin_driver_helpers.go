package logic

import (
	"context"
	"errors"
	"sync"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// buildAdminDriverPB 将司机基础资料、车辆、认证和 Redis 在线状态聚合为后台查询模型。
// driversvc 是司机域权威服务，adminsvc 只消费该聚合结果，不再直连司机域数据表。
// 车辆/认证/在线三条查询彼此独立，并行触发把单司机聚合耗时从"三倍单查询"压到接近单查询，
// 避免远端数据库延迟下因 adminsvc → driversvc RPC 超时导致整页列表 DeadlineExceeded。
func buildAdminDriverPB(ctx context.Context, svcCtx *svc.ServiceContext, d *model.Driver) (*proto.Driver, error) {
	if d == nil {
		return nil, repository.ErrDriverNotFound
	}
	item := &proto.Driver{
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
	}

	var (
		wg              sync.WaitGroup
		onlineState     *onlinestore.State
		vehicle         *model.DriverVehicle
		certification   *model.DriverCertification
		onlineErrCh     = make(chan error, 1)
		vehicleErrCh    = make(chan error, 1)
		certificationCh = make(chan error, 1)
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		if svcCtx.OnlineStore == nil {
			return
		}
		state, err := svcCtx.OnlineStore.Get(ctx, int64(d.Id))
		if err != nil {
			onlineErrCh <- err
			return
		}
		onlineState = state
	}()
	go func() {
		defer wg.Done()
		if svcCtx.DriverVehicleRepository == nil {
			return
		}
		v, err := svcCtx.DriverVehicleRepository.GetByDriverID(ctx, d.Id)
		if err != nil && !errors.Is(err, repository.ErrVehicleNotFound) {
			vehicleErrCh <- err
			return
		}
		vehicle = v
	}()
	go func() {
		defer wg.Done()
		if svcCtx.CertificationRepository == nil {
			return
		}
		c, err := svcCtx.CertificationRepository.GetByDriverID(ctx, d.Id)
		if err != nil && !errors.Is(err, repository.ErrCertificationNotFound) {
			certificationCh <- err
			return
		}
		certification = c
	}()
	wg.Wait()

	logger := logx.WithContext(ctx)
	if err, ok := <-onlineErrCh; ok {
		logger.Errorf("buildAdminDriverPB: get online state for driver %d failed: %v", d.Id, err)
	} else if onlineState != nil {
		item.OnlineStatus = onlineState.OnlineStatus
	} else {
		item.OnlineStatus = onlinestore.Offline
	}
	if err, ok := <-vehicleErrCh; ok {
		logger.Errorf("buildAdminDriverPB: get vehicle for driver %d failed: %v", d.Id, err)
	} else if vehicle != nil {
		item.VehicleId = int64(vehicle.Id)
		item.PlateNo = vehicle.PlateNo
		item.VehicleStatus = int32(vehicle.Status)
	}
	if err, ok := <-certificationCh; ok {
		logger.Errorf("buildAdminDriverPB: get certification for driver %d failed: %v", d.Id, err)
	} else if certification != nil {
		item.CertificationId = int64(certification.Id)
		item.AuditStatus = int32(certification.AuditStatus)
		item.AuditRemark = certification.AuditRemark
		if item.VehicleId == 0 {
			item.VehicleId = int64(certification.VehicleId)
		}
	}
	return item, nil
}
