package model

import "time"

// DriverPunishmentEffect 对应 driver_punishment_effect 表：管理端处罚动作的生效记录。
// 按 event_id + action_type 唯一，保证同一处罚事件的同一动作只生效一次（幂等）。
type DriverPunishmentEffect struct {
	Id                  uint64    `gorm:"primaryKey;column:id" json:"id"`
	EventId             string    `gorm:"column:event_id;size:64;not null" json:"eventId"`
	PunishmentNo        string    `gorm:"column:punishment_no;size:64;not null" json:"punishmentNo"`
	DriverId            uint64    `gorm:"column:driver_id;not null" json:"driverId"`
	ActionType          string    `gorm:"column:action_type;size:32;not null" json:"actionType"`
	ScoreDelta          int32     `gorm:"column:score_delta;default:0" json:"scoreDelta"`
	PriorityWeightDelta int32     `gorm:"column:priority_weight_delta;default:0" json:"priorityWeightDelta"`
	CreatedAt           time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName 返回对应的数据库表名。
func (DriverPunishmentEffect) TableName() string {
	return "driver_punishment_effect"
}
