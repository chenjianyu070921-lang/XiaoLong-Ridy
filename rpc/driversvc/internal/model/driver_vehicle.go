package model

import "time"

// DriverVehicle 对应 driver_vehicle 表：司机绑定的车辆信息。
type DriverVehicle struct {
	Id               uint64     `gorm:"primaryKey;column:id" json:"id"`                     // Id：车辆主键 ID（自增）
	DriverId         uint64     `gorm:"column:driver_id" json:"driverId"`                   // DriverId：所属司机 ID
	PlateNo          string     `gorm:"column:plate_no;size:20" json:"plateNo"`             // PlateNo：车牌号，唯一
	Brand            string     `gorm:"column:brand;size:50;default:''" json:"brand"`       // Brand：品牌
	Model            string     `gorm:"column:model;size:50;default:''" json:"model"`       // Model：车型
	Color            string     `gorm:"column:color;size:20;default:''" json:"color"`       // Color：车身颜色
	VehicleType      int8       `gorm:"column:vehicle_type;default:1" json:"vehicleType"`   // VehicleType：车辆类型；1特惠快车 2快车 3拼车
	RegistrationDate *time.Time `gorm:"column:registration_date" json:"registrationDate"`  // RegistrationDate：注册日期
	InsuranceNo      string     `gorm:"column:insurance_no;size:50;default:''" json:"insuranceNo"`       // InsuranceNo：保险单号
	InsuranceExpireAt *time.Time `gorm:"column:insurance_expire_at" json:"insuranceExpireAt"`          // InsuranceExpireAt：保险到期日
	Status           int8       `gorm:"column:status;default:1" json:"status"`             // Status：状态；1待审核 2正常 3禁用
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`                // CreatedAt：创建时间
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updatedAt"`                // UpdatedAt：更新时间
}

// TableName 返回对应的数据库表名。
func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}
