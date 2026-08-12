package model

import (
	"time"

	"gorm.io/gorm"
)

// Poi 高德 POI 搜索结果缓存表（servicecontext 启动时会自动建表）
type Poi struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;size:100;not null"`
	Address   string    `gorm:"column:address;size:255"`
	Longitude float64   `gorm:"column:longitude;type:decimal(10,6)"`
	Latitude  float64   `gorm:"column:latitude;type:decimal(10,6)"`
	Category  string    `gorm:"column:category;size:100"`
	Distance  int       `gorm:"column:distance"`
	Source    string    `gorm:"column:source;size:20;default:amap"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Poi) TableName() string {
	return "poi"
}

// PoiModel POI 数据访问层
type PoiModel struct {
	db *gorm.DB
}

func NewPoiModel(db *gorm.DB) *PoiModel {
	return &PoiModel{db: db}
}

// SearchByName 按关键词模糊查缓存，按距离排序，取前 limit 条
func (m *PoiModel) SearchByName(keyword string, limit int) ([]Poi, error) {
	if limit <= 0 {
		limit = 20
	}
	var list []Poi
	err := m.db.Where("name LIKE ?", "%"+keyword+"%").
		Order("distance ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// BatchInsert 批量写入缓存
func (m *PoiModel) BatchInsert(list []Poi) error {
	if len(list) == 0 {
		return nil
	}
	return m.db.Create(&list).Error
}
