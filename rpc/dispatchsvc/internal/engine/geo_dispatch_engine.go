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
		if e.enableMock {
			return NewMockDispatchEngine().FindCandidates(ctx, 0, fromLongitude, fromLatitude, 0, cityCode, normalizeOrderType(orderType))
		}
		return nil, nil
	}

	locs := make([]redis.GeoLocation, 0, len(byID))
	for _, loc := range byID {
		locs = append(locs, loc)
	}
	locs = e.filterAvailable(ctx, locs)
	locs = e.filterPreference(ctx, locs, orderType)
	if len(locs) == 0 {
		return nil, nil
	}
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
		score := distanceScore*0.6 + rating*10*0.3 + completion*100*0.1
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
