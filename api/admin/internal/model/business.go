// Package model 定义管理后台访问业务表时使用的数据模型。
package model

import "time"

// User 表示 user 表对应的乘客用户模型。
// 管理后台 P0 使用该模型完成用户列表和详情查询。
type User struct {
	ID             int64
	Phone          string
	PasswordHash   string
	Nickname       string
	AvatarURL      string
	Gender         int32
	RealName       string
	IDCardNo       string
	RegisterSource string
	Status         int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

// DriverCertification 表示司机认证审核聚合模型。
// 该结构聚合 driver_certification、driver 和 driver_vehicle 三张表的数据。
type DriverCertification struct {
	ID                int64
	DriverID          int64
	VehicleID         int64
	DriverPhone       string
	DriverName        string
	DriverStatus      int32
	PlateNo           string
	VehicleStatus     int32
	IDCardFrontURL    string
	IDCardBackURL     string
	DriverLicenseURL  string
	VehicleLicenseURL string
	AuditStatus       int32
	AuditRemark       string
	AuditedBy         int64
	AuditedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RideOrder 表示 ride_order 订单主表模型。
type RideOrder struct {
	ID                 int64
	OrderNo            string
	UserID             int64
	DriverID           int64
	CarType            int32
	FromAddress        string
	FromLongitude      string
	FromLatitude       string
	ToAddress          string
	ToLongitude        string
	ToLatitude         string
	EstimatedDistanceM int64
	EstimatedDurationS int64
	EstimatedPrice     string
	Status             int32
	CancelReason       string
	CancelBy           string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

// OperationLog 表示 admin_operation_log 后台操作日志模型。
type OperationLog struct {
	ID         int64
	AdminID    int64
	Module     string
	Action     string
	TargetType string
	TargetID   int64
	Detail     string
	IP         string
	CreatedAt  time.Time
}
