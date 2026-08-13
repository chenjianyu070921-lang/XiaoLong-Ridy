package logic

import (
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

// enumDriverStatus 将可选状态字符串转为 proto 枚举指针；nil 或空串返回 nil（不更新）。
func enumDriverStatus(s *string) *driversproto.DriverStatus {
	if s == nil || *s == "" {
		return nil
	}
	var v driversproto.DriverStatus
	switch *s {
	case "DRIVER_STATUS_PENDING":
		v = driversproto.DriverStatus_DRIVER_STATUS_PENDING
	case "DRIVER_STATUS_NORMAL":
		v = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	case "DRIVER_STATUS_FROZEN":
		v = driversproto.DriverStatus_DRIVER_STATUS_FROZEN
	case "DRIVER_STATUS_CANCELLED":
		v = driversproto.DriverStatus_DRIVER_STATUS_CANCELLED
	default:
		v = driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	return &v
}

// enumDriverStatusStr 将状态字符串转为 proto 枚举（用于列表过滤，空串表示不限）。
func enumDriverStatusStr(s string) driversproto.DriverStatus {
	if s == "" {
		return driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	switch s {
	case "DRIVER_STATUS_PENDING":
		return driversproto.DriverStatus_DRIVER_STATUS_PENDING
	case "DRIVER_STATUS_NORMAL":
		return driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	case "DRIVER_STATUS_FROZEN":
		return driversproto.DriverStatus_DRIVER_STATUS_FROZEN
	case "DRIVER_STATUS_CANCELLED":
		return driversproto.DriverStatus_DRIVER_STATUS_CANCELLED
	default:
		return driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
}
