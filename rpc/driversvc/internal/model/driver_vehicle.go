package model

import "time"

// DriverVehicle 对应 driver_vehicle 表：司机绑定的车辆信息。
type DriverVehicle struct {
	Id               uint64     `gorm:"primaryKey;column:id" json:"id"`
	DriverId         uint64     `gorm:"column:driver_id" json:"driverId"`
	PlateNo          string     `gorm:"column:plate_no;size:20" json:"plateNo"`
	Brand            string     `gorm:"column:brand;size:50;default:''" json:"brand"`
	Model            string     `gorm:"column:model;size:50;default:''" json:"model"`
	Color            string     `gorm:"column:color;size:20;default:''" json:"color"`
	VehicleType      int8       `gorm:"column:vehicle_type;default:1" json:"vehicleType"`
	RegistrationDate *time.Time `gorm:"column:registration_date" json:"registrationDate"`
	InsuranceNo      string     `gorm:"column:insurance_no;size:50;default:''" json:"insuranceNo"`
	InsuranceExpireAt *time.Time `gorm:"column:insurance_expire_at" json:"insuranceExpireAt"`
	Status           int8       `gorm:"column:status;default:1" json:"status"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName 返回对应的数据库表名。
func (DriverVehicle) TableName() string {
	return "driver_vehicle"
}
