package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/onlinestore"
	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

// buildAdminDriverPB 将司机基础资料、车辆、认证和 Redis 在线状态聚合为后台查询模型。
// driversvc 是司机域权威服务，adminsvc 只消费该聚合结果，不再直连司机域数据表。
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
	if svcCtx.OnlineStore != nil {
		state, err := svcCtx.OnlineStore.Get(ctx, int64(d.Id))
		if err != nil {
			return nil, err
		}
		if state == nil {
			item.OnlineStatus = onlinestore.Offline
		} else {
			item.OnlineStatus = state.OnlineStatus
		}
	}
	if svcCtx.DriverVehicleRepository != nil {
		vehicle, err := svcCtx.DriverVehicleRepository.GetByDriverID(ctx, d.Id)
		if err != nil && !errors.Is(err, repository.ErrVehicleNotFound) {
			return nil, err
		}
		if vehicle != nil {
			item.VehicleId = int64(vehicle.Id)
			item.PlateNo = vehicle.PlateNo
			item.VehicleStatus = int32(vehicle.Status)
		}
	}
	if svcCtx.CertificationRepository != nil {
		cert, err := svcCtx.CertificationRepository.GetByDriverID(ctx, d.Id)
		if err != nil && !errors.Is(err, repository.ErrCertificationNotFound) {
			return nil, err
		}
		if cert != nil {
			item.CertificationId = int64(cert.Id)
			item.AuditStatus = int32(cert.AuditStatus)
			item.AuditRemark = cert.AuditRemark
			if item.VehicleId == 0 {
				item.VehicleId = int64(cert.VehicleId)
			}
		}
	}
	return item, nil
}
