package logic

import (
	"context"
	"fmt"
	"strconv"

	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"

	"github.com/redis/go-redis/v9"
)

const defaultDispatchCity = "default"

func syncDispatchDriverOnline(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64, longitude, latitude float64) error {
	if svcCtx == nil || svcCtx.RedisClient == nil {
		return nil
	}
	member := strconv.FormatInt(driverID, 10)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	_, err := svcCtx.RedisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.GeoAdd(ctx, geoKey, &redis.GeoLocation{
			Name:      member,
			Longitude: longitude,
			Latitude:  latitude,
		})
		pipe.SAdd(ctx, constants.RedisDriverOnline, member)
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
	_, err := svcCtx.RedisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.SRem(ctx, constants.RedisDriverOnline, member)
		pipe.ZRem(ctx, geoKey, member)
		return nil
	})
	return err
}
