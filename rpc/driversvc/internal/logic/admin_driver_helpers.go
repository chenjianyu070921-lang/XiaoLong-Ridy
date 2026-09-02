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
// 车辆/认证/在线三条查询彼此独立，并行触发把单司机聚合耗时从"三倍单查询"压到接近单查询。
// 注意：三条 goroutine 各自写入独立的错误/结果变量，wg.Wait 的 happens-before 保证读取安全，
// 不要用"从空 buffered channel 接收"来传播错误（会永久阻塞）。
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
		wg            sync.WaitGroup
		onlineState   *onlinestore.State
		vehicle       *model.DriverVehicle
		certification *model.DriverCertification
		onlineErr     error
		vehicleErr    error
		certErr       error
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		if svcCtx.OnlineStore == nil {
			return
		}
		state, err := svcCtx.OnlineStore.Get(ctx, int64(d.Id))
		if err != nil {
			onlineErr = err
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
			vehicleErr = err
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
			certErr = err
			return
		}
		certification = c
	}()
	wg.Wait()
	close(onlineErrCh)
	close(vehicleErrCh)
	close(certificationCh)

	logger := logx.WithContext(ctx)
	if onlineErr != nil {
		logger.Errorf("buildAdminDriverPB: get online state for driver %d failed: %v", d.Id, onlineErr)
	} else if onlineState != nil {
		item.OnlineStatus = onlineState.OnlineStatus
	} else {
		item.OnlineStatus = onlinestore.Offline
	}
	if vehicleErr != nil {
		logger.Errorf("buildAdminDriverPB: get vehicle for driver %d failed: %v", d.Id, vehicleErr)
	} else if vehicle != nil {
		item.VehicleId = int64(vehicle.Id)
		item.PlateNo = vehicle.PlateNo
		item.VehicleStatus = int32(vehicle.Status)
	}
	if certErr != nil {
		logger.Errorf("buildAdminDriverPB: get certification for driver %d failed: %v", d.Id, certErr)
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
