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
