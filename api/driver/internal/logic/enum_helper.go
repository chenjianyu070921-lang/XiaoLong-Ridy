// Package logic 实现 driver API 的业务逻辑层。
package logic

import (
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的枚举定义来源
)

// enumDriverStatus 将可选状态字符串转为 proto 枚举指针；nil 或空串返回 nil（表示不更新该字段）。
func enumDriverStatus(s *string) *driversproto.DriverStatus {
	// 入参为 nil 或空串时返回 nil，调用方据此跳过该可选字段。
	if s == nil || *s == "" {
		return nil
	}
	// 声明局部变量承载映射后的枚举值。
	var v driversproto.DriverStatus
	// 按字符串值映射到对应的 proto 枚举。
	switch *s {
	case "DRIVER_STATUS_PENDING": // 待审核
		v = driversproto.DriverStatus_DRIVER_STATUS_PENDING
	case "DRIVER_STATUS_NORMAL": // 正常
		v = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	case "DRIVER_STATUS_FROZEN": // 冻结
		v = driversproto.DriverStatus_DRIVER_STATUS_FROZEN
	case "DRIVER_STATUS_CANCELLED": // 注销
		v = driversproto.DriverStatus_DRIVER_STATUS_CANCELLED
	default: // 未知值映射为未指定（由底层忽略）。
		v = driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	// 返回枚举值的指针，以匹配 proto 的 optional 字段语义。
	return &v
}

// enumDriverStatusStr 将状态字符串转为 proto 枚举（用于列表过滤，空串表示不限）。
func enumDriverStatusStr(s string) driversproto.DriverStatus {
	// 空串表示不过滤，返回未指定枚举。
	if s == "" {
		return driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	// 按字符串值映射到对应的 proto 枚举（与 enumDriverStatus 逻辑一致，但返回非指针）。
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
