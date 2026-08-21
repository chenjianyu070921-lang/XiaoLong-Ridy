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

// DriverScoreProvider 提供司机的真实服务质量评分，用于派单权重计算。
// 返回评分(rating, 0~5)与完单率(completion, 0~1)。查询失败或缺失时返回零值，引擎降级为默认权重。
type DriverScoreProvider func(ctx context.Context, driverID uint64) (rating float64, completion float64)

// geoDispatchEngine 基于 Redis GEO 的派单引擎：查附近司机并按距离/评分加权。
type geoDispatchEngine struct {
	rdb  *redis.Client
	city string
	// enableMock 是否允许在 GEO 查不到司机时回退 mock 候选，仅用于联调演示。
	enableMock bool
	// scoreProvider 真实司机评分查询，nil 时评分权重使用默认值。
	scoreProvider DriverScoreProvider
}

// NewGeoDispatchEngine 创建 Redis GEO 派单引擎。
func NewGeoDispatchEngine(rdb *redis.Client, city string) DispatchEngine {
	return NewGeoDispatchEngineWithMock(rdb, city, false)
}

// NewGeoDispatchEngineWithMock 创建 Redis GEO 派单引擎，并指定是否允许 mock 回退。
func NewGeoDispatchEngineWithMock(rdb *redis.Client, city string, enableMock bool) DispatchEngine {
	if city == "" {
		city = "default"
	}
	return &geoDispatchEngine{rdb: rdb, city: city, enableMock: enableMock}
}

// NewGeoDispatchEngineWithScore 创建 Redis GEO 派单引擎，并注入真实司机评分查询（driver_score 表），
// 替换写死的默认权重。不设置时保持旧默认行为，兼容已有测试。
func NewGeoDispatchEngineWithScore(rdb *redis.Client, city string, enableMock bool, p DriverScoreProvider) DispatchEngine {
	e := NewGeoDispatchEngineWithMock(rdb, city, enableMock).(*geoDispatchEngine)
	e.scoreProvider = p
	return e
}

// FindCandidates 同时检索默认城市与指定城市 GEO，避免城市键不一致导致查空。
func (e *geoDispatchEngine) FindCandidates(ctx context.Context, _ uint64, fromLongitude, fromLatitude float64, _ int32, cityCode string) ([]Candidate, error) {
	if e.rdb == nil {
		// 未配置 Redis 时视为无候选，不 panic。
		if e.enableMock {
			return NewMockDispatchEngine().FindCandidates(ctx, 0, fromLongitude, fromLatitude, 0, cityCode)
		}
		return nil, nil
	}

	keys := []string{fmt.Sprintf(constants.RedisDriverGeo, e.city)}
	if cityCode != "" && cityCode != e.city {
		keys = append(keys, fmt.Sprintf(constants.RedisDriverGeo, cityCode))
	}

	byID := make(map[uint64]redis.GeoLocation)
	for _, key := range keys {
		locs, err := e.rdb.GeoSearchLocation(ctx, key, &redis.GeoSearchLocationQuery{
			GeoSearchQuery: redis.GeoSearchQuery{
				Longitude:  fromLongitude,
				Latitude:   fromLatitude,
				Radius:     3000,
				RadiusUnit: "m",
				Sort:       "ASC",
				Count:      10,
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
	if len(byID) == 0 {
		// GEO 无数据时默认返回空候选（真实派单无司机可用）；
		// 仅当显式开启 EnableMockDispatch 时才回退 mock，避免联调"假成功"。
		if e.enableMock {
			return NewMockDispatchEngine().FindCandidates(ctx, 0, fromLongitude, fromLatitude, 0, cityCode)
		}
		return nil, nil
	}

	locs := make([]redis.GeoLocation, 0, len(byID))
	for _, loc := range byID {
		locs = append(locs, loc)
	}
	sort.Slice(locs, func(i, j int) bool { return locs[i].Dist < locs[j].Dist })

	candidates := make([]Candidate, 0, len(locs))
	for _, loc := range locs {
		distanceScore := math.Max(0, 100-loc.Dist/1000*10)
		rating, completion := 4.5, 0.9 // 默认权重，无真实评分数据时降级
		if e.scoreProvider != nil {
			if r, c := e.scoreProvider(ctx, mustParseID(loc.Name)); r > 0 || c > 0 {
				rating, completion = r, c
			}
		}
		score := distanceScore*0.6 + rating*10*0.3 + completion*100*0.1
		candidates = append(candidates, Candidate{DriverID: mustParseID(loc.Name), MatchScore: score})
	}
	return candidates, nil
}

func mustParseID(name string) uint64 {
	id, err := strconv.ParseUint(name, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
