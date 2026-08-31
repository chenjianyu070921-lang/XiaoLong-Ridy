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

	if err := rdb.SAdd(ctx, constants.RedisDriverBusy, "25").Err(); err != nil {
		t.Fatalf("SAdd busy() error = %v", err)
	}
	if err := syncDispatchDriverOnline(ctx, svcCtx, 25, 116.397, 39.908); err != nil {
		t.Fatalf("syncDispatchDriverOnline() error = %v", err)
	}
	// sync 只同步 geo/online/pos，绝不清除 busy——服务中司机依赖 busy 防止重复派单。
	busy, err := rdb.SIsMember(ctx, constants.RedisDriverBusy, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember() busy error = %v", err)
	}
	if !busy {
		t.Fatalf("syncDispatchDriverOnline must NOT clear busy (service-in-progress driver must stay busy)")
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

	// 上线入口负责清 busy：clearDispatchDriverBusy 只在上线时调用。
	if err := clearDispatchDriverBusy(ctx, svcCtx, 25); err != nil {
		t.Fatalf("clearDispatchDriverBusy() error = %v", err)
	}
	busy, err = rdb.SIsMember(ctx, constants.RedisDriverBusy, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember() busy after clear error = %v", err)
	}
	if busy {
		t.Fatalf("clearDispatchDriverBusy should remove driver from %s", constants.RedisDriverBusy)
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

func TestSyncDispatchDriverOnlineWritesPositionSnapshot(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{RedisClient: rdb}
	ctx := context.Background()

	if err := syncDispatchDriverOnline(ctx, svcCtx, 25, 116.397, 39.908); err != nil {
		t.Fatalf("syncDispatchDriverOnline() error = %v", err)
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

func TestSyncDispatchDriverOnlineRejectsZeroCoordinatePair(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	svcCtx := &svc.ServiceContext{RedisClient: rdb}
	ctx := context.Background()

	if err := syncDispatchDriverOnline(ctx, svcCtx, 25, 0, 0); err == nil {
		t.Fatal("syncDispatchDriverOnline() accepted empty decoded coordinates 0,0")
	}

	posKey := fmt.Sprintf(constants.RedisDriverPos, 25)
	exists, err := rdb.Exists(ctx, posKey).Result()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists != 0 {
		t.Fatalf("position snapshot should not be written for 0,0, exists=%d", exists)
	}
	online, err := rdb.SIsMember(ctx, constants.RedisDriverOnline, "25").Result()
	if err != nil {
		t.Fatalf("SIsMember() error = %v", err)
	}
	if online {
		t.Fatalf("driver should not be added to %s for 0,0", constants.RedisDriverOnline)
	}
}
