package model

import "time"

// DriverCertification 对应 driver_certification 表：司机/车辆资质上传与审核记录。
type DriverCertification struct {
	Id                uint64     `gorm:"primaryKey;column:id" json:"id"`                          // Id：认证主键 ID（自增）
	DriverId          uint64     `gorm:"column:driver_id" json:"driverId"`                        // DriverId：所属司机 ID
	VehicleId         uint64     `gorm:"column:vehicle_id;default:0" json:"vehicleId"`           // VehicleId：关联车辆 ID（可空）
	IdCardFrontUrl    string     `gorm:"column:id_card_front_url;size:255;default:''" json:"idCardFrontUrl"`   // IdCardFrontUrl：身份证人像面
	IdCardBackUrl     string     `gorm:"column:id_card_back_url;size:255;default:''" json:"idCardBackUrl"`     // IdCardBackUrl：身份证国徽面
	DriverLicenseUrl  string     `gorm:"column:driver_license_url;size:255;default:''" json:"driverLicenseUrl"` // DriverLicenseUrl：驾驶证照片
	VehicleLicenseUrl string     `gorm:"column:vehicle_license_url;size:255;default:''" json:"vehicleLicenseUrl"` // VehicleLicenseUrl：行驶证照片
	AuditStatus       int8       `gorm:"column:audit_status;default:1" json:"auditStatus"`       // AuditStatus：审核状态；1待审核 2通过 3驳回
	AuditRemark       string     `gorm:"column:audit_remark;size:255;default:''" json:"auditRemark"` // AuditRemark：驳回原因/审核备注
	AuditedBy         uint64     `gorm:"column:audited_by;default:0" json:"auditedBy"`            // AuditedBy：审核人（管理员 ID）
	AuditedAt         *time.Time `gorm:"column:audited_at" json:"auditedAt"`                      // AuditedAt：审核时间
	CreatedAt         time.Time  `gorm:"column:created_at" json:"createdAt"`                     // CreatedAt：创建时间
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updatedAt"`                     // UpdatedAt：更新时间
}

// TableName 返回对应的数据库表名。
func (DriverCertification) TableName() string {
	return "driver_certification"
}
