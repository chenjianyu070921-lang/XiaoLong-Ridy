// Package onlinestore 封装司机在线状态在 Redis 中的读写与多端互踢判定逻辑。
// 在线状态以 Redis 为权威（TTL 保活），MySQL driver_location 仅做持久化兜底与派单查询。
package onlinestore

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// 在线状态常量，与 driver_location.online_status 语义一致。
const (
	// Offline 表示司机离线。
	Offline int32 = 0
	// Online 表示司机在线（可接单）。
	Online int32 = 1
	// OnTrip 表示司机行程中。
	OnTrip int32 = 2
)

// defaultTTL 在线状态默认存活时长（心跳超时阈值），超过即视为离线。
const defaultTTL = 60 * time.Second

// keyPrefix 在线状态 Hash 的 key 前缀。
const keyPrefix = "driver:online:"

// Store 提供司机在线状态在 Redis 中的增改查与互踢判定。
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore 创建基于 go-redis 的在线状态存储，ttl<=0 时使用默认 60s。
func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Store{rdb: rdb, ttl: ttl}
}

// State 描述司机在线状态快照。
type State struct {
	OnlineStatus int32   // 在线状态
	DeviceID     string  // 当前绑定设备标识
	Longitude    float64 // 经度
	Latitude     float64 // 纬度
}

// heartbeatable 表示心跳后是否需要续期 TTL（在线时续期，离线不续期）。
func (s *Store) key(driverID int64) string {
	return keyPrefix + strconv.FormatInt(driverID, 10)
}

// SetOnline 记录司机上线：写入在线状态与设备标识，并重置 TTL。
func (s *Store) SetOnline(ctx context.Context, driverID int64, deviceID string, longitude, latitude float64) error {
	key := s.key(driverID)
	_, err := s.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, map[string]interface{}{
			"online_status": Online,
			"device_id":     deviceID,
			"longitude":     longitude,
			"latitude":      latitude,
		})
		pipe.Expire(ctx, key, s.ttl)
		return nil
	})
	return err
}

// SetOffline 记录司机下线：仅置离线状态（保留设备标识供互踢判定），不清除 key 以便弱化离线状态。
func (s *Store) SetOffline(ctx context.Context, driverID int64) error {
	key := s.key(driverID)
	return s.rdb.HSet(ctx, key, "online_status", Offline).Err()
}

// SetStatus 更新司机服务状态，保留已有设备与位置；在线/行程中状态会续期 TTL。
func (s *Store) SetStatus(ctx context.Context, driverID int64, onlineStatus int32) error {
	key := s.key(driverID)
	_, err := s.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, "online_status", onlineStatus)
		if onlineStatus != Offline {
			pipe.Expire(ctx, key, s.ttl)
		}
		return nil
	})
	return err
}

// Get 读取司机当前在线状态；不存在时返回 (nil, nil)。
func (s *Store) Get(ctx context.Context, driverID int64) (*State, error) {
	vals, err := s.rdb.HGetAll(ctx, s.key(driverID)).Result()
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	st, err := strconv.ParseInt(vals["online_status"], 10, 32)
	if err != nil {
		st = int64(Offline)
	}
	lon, _ := strconv.ParseFloat(vals["longitude"], 64)
	lat, _ := strconv.ParseFloat(vals["latitude"], 64)
	return &State{
		OnlineStatus: int32(st),
		DeviceID:     vals["device_id"],
		Longitude:    lon,
		Latitude:     lat,
	}, nil
}

// Heartbeat 处理心跳：校验设备是否仍为当前绑定设备。
// 若 deviceID 与已绑定设备不一致（被新设备登录顶替）返回 kicked=true；
// 在线状态的心跳会重置 TTL 实现保活。
func (s *Store) Heartbeat(ctx context.Context, driverID int64, deviceID string, longitude, latitude float64) (onlineStatus int32, kicked bool, err error) {
	key := s.key(driverID)
	// 读取当前在线状态与绑定设备。
	st, err := s.Get(ctx, driverID)
	if err != nil {
		return Offline, false, err
	}
	// 无在线记录视为离线，直接写心跳为在线并绑定设备。
	if st == nil {
		if err := s.SetOnline(ctx, driverID, deviceID, longitude, latitude); err != nil {
			return Offline, false, err
		}
		return Online, false, nil
	}
	// 设备不匹配：判定被顶替，踢出当前设备（返回 kicked，不覆盖已有在线状态）。
	if st.DeviceID != "" && st.DeviceID != deviceID {
		return st.OnlineStatus, true, nil
	}
	// 设备匹配：更新位置并续期 TTL。
	_, err = s.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, key, "longitude", longitude, "latitude", latitude)
		pipe.Expire(ctx, key, s.ttl)
		return nil
	})
	if err != nil {
		return st.OnlineStatus, false, err
	}
	return st.OnlineStatus, false, nil
}
