package logic

import (
	"context"
	"fmt"
	"testing"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestSyncDispatchDriverAvailability(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{RedisClient: rdb}
	ctx := context.Background()

	if err := syncDispatchDriverOnline(ctx, svcCtx, 25, 116.397, 39.908); err != nil {
		t.Fatalf("syncDispatchDriverOnline() error = %v", err)
	}
	online, err := rdb.SIsMember(ctx, constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember() error = %v", err)
	}
	if !online {
		t.Fatalf("driver should be in %s", constants.RedisDriverOnline)
	}
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	locations, err := rdb.GeoPos(ctx, geoKey, "25").Result()
	if err != nil {
		t.Fatalf("GeoPos() error = %v", err)
	}
	if len(locations) != 1 || locations[0] == nil {
		t.Fatalf("driver geo position missing: %+v", locations)
	}

	if err := syncDispatchDriverOffline(ctx, svcCtx, 25); err != nil {
		t.Fatalf("syncDispatchDriverOffline() error = %v", err)
	}
	online, err = rdb.SIsMember(ctx, constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember() after offline error = %v", err)
	}
	if online {
		t.Fatalf("driver should be removed from %s", constants.RedisDriverOnline)
	}
	locations, err = rdb.GeoPos(ctx, geoKey, "25").Result()
	if err != nil {
		t.Fatalf("GeoPos() after offline error = %v", err)
	}
	if len(locations) != 1 || locations[0] != nil {
		t.Fatalf("driver geo position should be removed: %+v", locations)
	}
}

func TestSyncDispatchDriverOnlinePreferenceSets(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{RedisClient: rdb}
	ctx := context.Background()

	pref := driverListenPreference{AcceptRealtime: false, AcceptReservation: true}
	if err := syncDispatchDriverOnlineWithPreference(ctx, svcCtx, 25, 116.397, 39.908, pref); err != nil {
		t.Fatalf("syncDispatchDriverOnlineWithPreference() error = %v", err)
	}
	realtime, err := rdb.SIsMember(ctx, constants.RedisDriverPrefRealtime, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember realtime error = %v", err)
	}
	reservation, err := rdb.SIsMember(ctx, constants.RedisDriverPrefReservation, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember reservation error = %v", err)
	}
	if realtime || !reservation {
		t.Fatalf("preference sets realtime=%v reservation=%v, want false/true", realtime, reservation)
	}

	if err := syncDispatchDriverOffline(ctx, svcCtx, 25); err != nil {
		t.Fatalf("syncDispatchDriverOffline() error = %v", err)
	}
	reservation, err = rdb.SIsMember(ctx, constants.RedisDriverPrefReservation, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember reservation after offline error = %v", err)
	}
	if reservation {
		t.Fatal("offline driver should be removed from reservation preference set")
	}
}
func TestSyncDispatchDriverOnlineWritesPositionSnapshot(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{RedisClient: rdb}
	ctx := context.Background()

	pref := driverListenPreference{AcceptRealtime: true, AcceptReservation: true}
	if err := syncDispatchDriverOnlineWithPreference(ctx, svcCtx, 25, 116.397, 39.908, pref); err != nil {
		t.Fatalf("syncDispatchDriverOnlineWithPreference() error = %v", err)
	}

	posKey := fmt.Sprintf(constants.RedisDriverPos, 25)
	pos, err := rdb.HGetAll(ctx, posKey).Result()
	if err != nil {
		t.Fatalf("HGetAll() error = %v", err)
	}
	if pos["driver_id"] != "25" || pos["longitude"] != "116.397000" || pos["latitude"] != "39.908000" || pos["report_time"] == "" {
		t.Fatalf("position snapshot = %+v", pos)
	}

	if err := syncDispatchDriverOffline(ctx, svcCtx, 25); err != nil {
		t.Fatalf("syncDispatchDriverOffline() error = %v", err)
	}
	exists, err := rdb.Exists(ctx, posKey).Result()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists != 0 {
		t.Fatalf("position snapshot should be removed, exists=%d", exists)
	}
}
