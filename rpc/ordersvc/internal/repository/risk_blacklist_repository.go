package repository

import (
	"context"
	"strconv"
	"sync"

	"gorm.io/gorm"
)

// BlacklistEntry 表示订单服务处理黑名单命中所需的最小字段集。
type BlacklistEntry struct {
	ID     uint64
	Reason string
}

// BlacklistHitRecord 表示一次下单黑名单命中审计记录。
type BlacklistHitRecord struct {
	BlacklistID uint64
	TargetType  string
	TargetID    uint64
	Scene       string
	RiskLevel   int8
	HitReason   string
	RequestID   string
}

// RiskBlacklistRepository 定义订单服务查询黑名单并保存命中记录的依赖契约。
type RiskBlacklistRepository interface {
	FindActiveByTarget(ctx context.Context, targetType string, targetID uint64) (*BlacklistEntry, error)
	CreateHitRecord(ctx context.Context, record *BlacklistHitRecord) error
}

// gormRiskBlacklistRepository 使用订单服务业务库访问运营风控表。
type gormRiskBlacklistRepository struct {
	db *gorm.DB
}

// NewGormRiskBlacklistRepository 创建生产环境使用的黑名单仓储。
func NewGormRiskBlacklistRepository(db *gorm.DB) RiskBlacklistRepository {
	return &gormRiskBlacklistRepository{db: db}
}

// FindActiveByTarget 查询指定对象当前生效的最新黑名单记录。
func (r *gormRiskBlacklistRepository) FindActiveByTarget(ctx context.Context, targetType string, targetID uint64) (*BlacklistEntry, error) {
	var entry BlacklistEntry
	err := r.db.WithContext(ctx).Table("blacklist").
		Select("id, reason").
		Where("target_type = ? AND target_id = ? AND status = 1", targetType, targetID).
		Order("id DESC").First(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// CreateHitRecord 持久化命中审计记录，写入失败必须反馈给下单入口处理。
func (r *gormRiskBlacklistRepository) CreateHitRecord(ctx context.Context, record *BlacklistHitRecord) error {
	return r.db.WithContext(ctx).Table("risk_blacklist_hit_record").Create(map[string]interface{}{
		"blacklist_id": record.BlacklistID,
		"target_type":  record.TargetType,
		"target_id":    record.TargetID,
		"scene":        record.Scene,
		"risk_level":   record.RiskLevel,
		"hit_reason":   record.HitReason,
		"request_id":   record.RequestID,
	}).Error
}

// MemoryRiskBlacklistRepository 为订单单元测试提供线程安全的内存实现。
type MemoryRiskBlacklistRepository struct {
	mu      sync.RWMutex
	entries map[string]*BlacklistEntry
	Hits    []BlacklistHitRecord
	Err     error
}

// NewMemoryRiskBlacklistRepository 创建空的内存黑名单仓储。
func NewMemoryRiskBlacklistRepository() *MemoryRiskBlacklistRepository {
	return &MemoryRiskBlacklistRepository{entries: make(map[string]*BlacklistEntry)}
}

// SetActive 设置指定目标的生效黑名单，供测试构造命中场景。
func (r *MemoryRiskBlacklistRepository) SetActive(targetType string, targetID uint64, entry BlacklistEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[riskTargetKey(targetType, targetID)] = &entry
}

// FindActiveByTarget 查询内存中的生效黑名单。
func (r *MemoryRiskBlacklistRepository) FindActiveByTarget(_ context.Context, targetType string, targetID uint64) (*BlacklistEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.Err != nil {
		return nil, r.Err
	}
	entry := r.entries[riskTargetKey(targetType, targetID)]
	if entry == nil {
		return nil, nil
	}
	copied := *entry
	return &copied, nil
}

// CreateHitRecord 追加命中记录副本，便于测试断言写入字段。
func (r *MemoryRiskBlacklistRepository) CreateHitRecord(_ context.Context, record *BlacklistHitRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Err != nil {
		return r.Err
	}
	r.Hits = append(r.Hits, *record)
	return nil
}

// riskTargetKey 生成内存索引键，避免不同类型的相同目标 ID 相互覆盖。
func riskTargetKey(targetType string, targetID uint64) string {
	return targetType + ":" + strconv.FormatUint(targetID, 10)
}
