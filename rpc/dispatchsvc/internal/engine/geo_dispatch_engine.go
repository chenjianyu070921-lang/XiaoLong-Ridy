package engine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

	"XiaoLong-Ridy/common/constants"

	"github.com/redis/go-redis/v9"
)

const (
	OrderTypeRealtime    = int32(constants.OrderTypeRealtime)
	OrderTypeReservation = int32(constants.OrderTypeReservation)
)

// dispatchRadiusLevels 派单半径降级梯度（单位：米）。
// 从近到远逐级尝试，优先匹配近处司机，查不到再扩大范围。
// 20km 兜底覆盖绝大多数城市主城区，极端情况（偏远新区）仍查不到则返回空。
var dispatchRadiusLevels = []int{3000, 5000, 10000, 20000}

// dispatchMaxCandidates 每级半径最多取多少个司机做筛选。
const dispatchMaxCandidates = 10

// 匹配分权重：距离分 0.6 + 评分分 0.3 + 完成率分 0.1。
// 就近派单优先，服务质量次之。
const (
	weightDistance   = 0.6
	weightRating     = 0.3
	weightCompletion = 0.1
)

type DriverScoreProvider func(ctx context.Context, driverID uint64) (rating float64, completion float64)

type DriverAvailability func(ctx context.Context, driverID uint64) (online, busy bool)

type DriverPreference func(ctx context.Context, driverID uint64, orderType int32) bool

type geoDispatchEngine struct {
	rdb           *redis.Client
	city          string
	enableMock    bool
	scoreProvider DriverScoreProvider
	availability  DriverAvailability
	preference    DriverPreference
}

func NewGeoDispatchEngine(rdb *redis.Client, city string) DispatchEngine {
	return NewGeoDispatchEngineWithMock(rdb, city, false)
}

func NewGeoDispatchEngineWithMock(rdb *redis.Client, city string, enableMock bool) DispatchEngine {
	if city == "" {
		city = "default"
	}
	return &geoDispatchEngine{rdb: rdb, city: city, enableMock: enableMock}
}

func NewGeoDispatchEngineWithScore(rdb *redis.Client, city string, enableMock bool, p DriverScoreProvider) DispatchEngine {
	e := NewGeoDispatchEngineWithMock(rdb, city, enableMock).(*geoDispatchEngine)
	e.scoreProvider = p
	return e
}

func NewGeoDispatchEngineWithScoreAndAvailability(rdb *redis.Client, city string, enableMock bool, p DriverScoreProvider, a DriverAvailability) DispatchEngine {
	return NewGeoDispatchEngineWithScoreAvailabilityAndPreference(rdb, city, enableMock, p, a, nil)
}

func NewGeoDispatchEngineWithScoreAvailabilityAndPreference(rdb *redis.Client, city string, enableMock bool, p DriverScoreProvider, a DriverAvailability, pref DriverPreference) DispatchEngine {
	e := NewGeoDispatchEngineWithScore(rdb, city, enableMock, p).(*geoDispatchEngine)
	e.availability = a
	e.preference = pref
	return e
}

func (e *geoDispatchEngine) FindCandidates(ctx context.Context, _ uint64, fromLongitude, fromLatitude float64, _ int32, cityCode string, orderType int32) ([]Candidate, error) {
	if e.rdb == nil {
		if e.enableMock {
			return NewMockDispatchEngine().FindCandidates(ctx, 0, fromLongitude, fromLatitude, 0, cityCode, normalizeOrderType(orderType))
		}
		return nil, nil
	}

	keys := buildGeoKeys(e.city, cityCode)

	// 多级半径降级：3km → 5km → 10km → 20km。
	// 每一级搜完司机后，先做可用性（在线+空闲）和偏好过滤，
	// 过滤后如果还有可用司机就直接返回，否则扩大半径继续找。
	// 这样既保证了近处优先，又避免了"小地方派不到单"的死锁。
	for _, radius := range dispatchRadiusLevels {
		locs, err := e.searchDriversInRadius(ctx, keys, fromLongitude, fromLatitude, radius)
		if err != nil {
			return nil, err
		}
		if len(locs) == 0 {
			continue
		}

		locs = e.filterAvailable(ctx, locs)
		locs = e.filterPreference(ctx, locs, orderType)
		if len(locs) > 0 {
			return e.scoreCandidates(ctx, locs)
		}
		// 这个半径内的司机都在忙 / 不接这个类型，扩大范围重试
	}

	// 全部半径都找不到 → mock 兜底（P0 联调兼容）
	if e.enableMock {
		return NewMockDispatchEngine().FindCandidates(ctx, 0, fromLongitude, fromLatitude, 0, cityCode, normalizeOrderType(orderType))
	}
	return nil, nil
}

// buildGeoKeys 根据城市码构造 Redis GEO 查询 key 列表。
// 额外追加 default key 保证跨城市配置不一致时也能搜到司机。
func buildGeoKeys(city, cityCode string) []string {
	keys := []string{fmt.Sprintf(constants.RedisDriverGeo, city)}
	if cityCode != "" && cityCode != city {
		keys = append(keys, fmt.Sprintf(constants.RedisDriverGeo, cityCode))
	}
	defaultKey := fmt.Sprintf(constants.RedisDriverGeo, "default")
	if defaultKey != keys[0] {
		dup := false
		for _, k := range keys {
			if k == defaultKey {
				dup = true
				break
			}
		}
		if !dup {
			keys = append(keys, defaultKey)
		}
	}
	return keys
}

// searchDriversInRadius 在指定半径内（单位：米）从多个城市 key 搜索司机，按去重+最近距离合并。
func (e *geoDispatchEngine) searchDriversInRadius(ctx context.Context, keys []string, lng, lat float64, radiusMeters int) ([]redis.GeoLocation, error) {
	byID := make(map[uint64]redis.GeoLocation)
	for _, key := range keys {
		locs, err := e.rdb.GeoSearchLocation(ctx, key, &redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  lng,
				Latitude:   lat,
				Radius:     float64(radiusMeters),
				RadiusUnit: "m",
				Sort:       "ASC",
				Count:      dispatchMaxCandidates,
			},
			WithDist: true,
		}).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}
		for _, loc := range locs {
			driverID, parseErr := strconv.ParseUint(loc.Name, 10, 64)
			if parseErr != nil {
				continue
			}
			if old, exists := byID[driverID]; !exists || loc.Dist < old.Dist {
				byID[driverID] = loc
			}
		}
	}
	locs := make([]redis.GeoLocation, 0, len(byID))
	for _, loc := range byID {
		locs = append(locs, loc)
	}
	return locs, nil
}

// scoreCandidates 对筛选后的司机按距离排序并计算匹配分。
func (e *geoDispatchEngine) scoreCandidates(ctx context.Context, locs []redis.GeoLocation) ([]Candidate, error) {
	sort.Slice(locs, func(i, j int) bool { return locs[i].Dist < locs[j].Dist })

	candidates := make([]Candidate, 0, len(locs))
	for _, loc := range locs {
		distanceScore := math.Max(0, 100-loc.Dist/1000*10)
		rating, completion := 4.5, 0.9
		if e.scoreProvider != nil {
			if r, c := e.scoreProvider(ctx, mustParseID(loc.Name)); r > 0 || c > 0 {
				rating, completion = r, c
			}
		}
		score := distanceScore*weightDistance + rating*10*weightRating + completion*100*weightCompletion
		candidates = append(candidates, Candidate{DriverID: mustParseID(loc.Name), MatchScore: score})
	}
	return candidates, nil
}

func (e *geoDispatchEngine) filterAvailable(ctx context.Context, locs []redis.GeoLocation) []redis.GeoLocation {
	if e.availability == nil {
		return locs
	}
	filtered := locs[:0]
	for _, loc := range locs {
		online, busy := e.availability(ctx, mustParseID(loc.Name))
		if !online || busy {
			continue
		}
		filtered = append(filtered, loc)
	}
	return filtered
}

func (e *geoDispatchEngine) filterPreference(ctx context.Context, locs []redis.GeoLocation, orderType int32) []redis.GeoLocation {
	if e.preference == nil {
		return locs
	}
	orderType = normalizeOrderType(orderType)
	filtered := locs[:0]
	for _, loc := range locs {
		if e.preference(ctx, mustParseID(loc.Name), orderType) {
			filtered = append(filtered, loc)
		}
	}
	return filtered
}

func normalizeOrderType(orderType int32) int32 {
	if orderType == OrderTypeReservation {
		return OrderTypeReservation
	}
	return OrderTypeRealtime
}

func mustParseID(name string) uint64 {
	id, err := strconv.ParseUint(name, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
