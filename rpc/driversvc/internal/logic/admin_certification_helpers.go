package logic

import (
	"time"

	"XiaoLong-Ridy/rpc/driversvc/internal/repository"
	"XiaoLong-Ridy/rpc/driversvc/proto"
)

// toAdminCertification 将仓储关联行转换为管理后台协议结构。
func toAdminCertification(row *repository.AdminCertificationRow) *proto.AdminCertification {
	if row == nil {
		return nil
	}
	return &proto.AdminCertification{
		Id:                int64(row.Id),
		DriverId:          int64(row.DriverId),
		VehicleId:         int64(row.VehicleId),
		DriverPhone:       row.DriverPhone,
		DriverName:        row.DriverName,
		DriverStatus:      int32(row.DriverStatus),
		PlateNo:           row.PlateNo,
		VehicleStatus:     int32(row.VehicleStatus),
		IdCardFrontUrl:    row.IdCardFrontUrl,
		IdCardBackUrl:     row.IdCardBackUrl,
		DriverLicenseUrl:  row.DriverLicenseUrl,
		VehicleLicenseUrl: row.VehicleLicenseUrl,
		AuditStatus:       int32(row.AuditStatus),
		AuditRemark:       row.AuditRemark,
		AuditedBy:         int64(row.AuditedBy),
		AuditedAt:         unixOrZero(row.AuditedAt),
		CreatedAt:         row.CreatedAt.Unix(),
		UpdatedAt:         row.UpdatedAt.Unix(),
	}
}

// unixOrZero 将可选时间转换为 Unix 秒，nil 返回 0。
func unixOrZero(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}
