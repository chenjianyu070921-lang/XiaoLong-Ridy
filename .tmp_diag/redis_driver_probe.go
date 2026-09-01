package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"XiaoLong-Ridy/common/constants"

	"github.com/redis/go-redis/v9"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run .tmp_diag/redis_driver_probe.go <driver_id>")
		os.Exit(2)
	}
	driverID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil || driverID <= 0 {
		fmt.Fprintln(os.Stderr, "invalid driver id")
		os.Exit(2)
	}

	addr := getenv("DRIVER_REDIS_ADDR", "115.191.16.159:6379")
	password := getenv("DRIVER_REDIS_PASSWORD", "4ay1nkal3u8ed77y")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0})
	defer rdb.Close()

	member := strconv.FormatInt(driverID, 10)
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, "default")

	pos, posErr := rdb.HGetAll(ctx, posKey).Result()
	online, onlineErr := rdb.SIsMember(ctx, constants.RedisDriverOnline, member).Result()
	geo, geoErr := rdb.GeoPos(ctx, geoKey, member).Result()
	ttl, ttlErr := rdb.TTL(ctx, posKey).Result()

	out := map[string]any{
		"addr":       addr,
		"driverId":   driverID,
		"posKey":     posKey,
		"pos":        pos,
		"onlineKey":  constants.RedisDriverOnline,
		"online":     online,
		"geoKey":     geoKey,
		"geo":        geo,
		"ttlSeconds": int64(ttl.Seconds()),
		"errors": map[string]string{
			"pos":    errString(posErr),
			"online": errString(onlineErr),
			"geo":    errString(geoErr),
			"ttl":    errString(ttlErr),
		},
	}
	data, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(data))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
