package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"

	"github.com/redis/go-redis/v9"
)

const defaultDispatchCity = "default"

const dispatchDriverPositionTTL = 2 * time.Minute

func syncDispatchDriverOnline(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, longitude, latitude float64) error {
	if svcCtx == nil || svcCtx.RedisClient == nil {
		return nil
	}
	member := strconv.FormatInt(driverID, 10)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	reportTime := strconv.FormatInt(time.Now().Unix(), 10)
	_, err := svcCtx.RedisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, constants.RedisDriverBusy, member)
		pipe.GeoAdd(ctx, geoKey, &redis.GeoLocation{
			Name:      member,
			Longitude: longitude,
			Latitude:  latitude,
		})
		pipe.SAdd(ctx, constants.RedisDriverOnline, member)
		pipe.HSet(ctx, posKey, map[string]interface{}{
			"driver_id":   member,
			"longitude":   strconv.FormatFloat(longitude, 'f', 6, 64),
			"latitude":    strconv.FormatFloat(latitude, 'f', 6, 64),
			"report_time": reportTime,
		})
		pipe.Expire(ctx, posKey, dispatchDriverPositionTTL)
		return nil
	})
	return err
}

func syncDispatchDriverOffline(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64) error {
	if svcCtx == nil || svcCtx.RedisClient == nil {
		return nil
	}
	member := strconv.FormatInt(driverID, 10)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	_, err := svcCtx.RedisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, constants.RedisDriverOnline, member)
		pipe.ZRem(ctx, geoKey, member)
		pipe.Del(ctx, posKey)
		return nil
	})
	return err
}
