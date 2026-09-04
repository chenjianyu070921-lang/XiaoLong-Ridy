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
	if driverID <= 0 || !validLongitudeLatitude(longitude, latitude) {
		return errInvalidLongitudeLatitude
	}
	member := strconv.FormatInt(driverID, 10)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	reportTime := strconv.FormatInt(time.Now().Unix(), 10)
	// 逐条写入：当前环境的 Redis 兼容层对 Pipelined 中的 SADD/ZADD 会静默丢弃（HSET 正常），
	// 若用 Pipelined 会导致司机 GEO/在线集合写不进去，派单引擎筛不到司机、司机端搜不到订单。
	if err := svcCtx.RedisClient.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      member,
		Longitude: longitude,
		Latitude:  latitude,
	}).Err(); err != nil {
		return err
	}
	if err := svcCtx.RedisClient.SAdd(ctx, constants.RedisDriverOnline, member).Err(); err != nil {
		return err
	}
	if err := svcCtx.RedisClient.HSet(ctx, posKey, map[string]interface{}{
		"driver_id":   member,
		"longitude":   strconv.FormatFloat(longitude, 'f', 6, 64),
		"latitude":    strconv.FormatFloat(latitude, 'f', 6, 64),
		"report_time": reportTime,
	}).Err(); err != nil {
		return err
	}
	return svcCtx.RedisClient.Expire(ctx, posKey, dispatchDriverPositionTTL).Err()
}

// clearDispatchDriverBusy 将司机从忙碌集合（driver:busy）中移除。
// 仅在司机【上线】时调用：用于清理异常订单/服务重启残留的 busy 标记，避免司机无法进入派单池。
// 注意：不能在心跳或位置上报时调用——服务中的司机需要保持 busy 状态防止重复派单（P0-回归修复）。
func clearDispatchDriverBusy(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64) error {
	if svcCtx == nil || svcCtx.RedisClient == nil {
		return nil
	}
	member := strconv.FormatInt(driverID, 10)
	return svcCtx.RedisClient.SRem(ctx, constants.RedisDriverBusy, member).Err()
}

func syncDispatchDriverOffline(ctx context.Context, svcCtx *svc.ServiceContext, driverID int64) error {
	if svcCtx == nil || svcCtx.RedisClient == nil {
		return nil
	}
	member := strconv.FormatInt(driverID, 10)
	geoKey := fmt.Sprintf(constants.RedisDriverGeo, defaultDispatchCity)
	posKey := fmt.Sprintf(constants.RedisDriverPos, driverID)
	// 逐条执行（同 syncDispatchDriverOnline：兼容层 Pipelined 会丢弃 ZREM/SREM）
	if err := svcCtx.RedisClient.SRem(ctx, constants.RedisDriverOnline, member).Err(); err != nil {
		return err
	}
	if err := svcCtx.RedisClient.ZRem(ctx, geoKey, member).Err(); err != nil {
		return err
	}
	return svcCtx.RedisClient.Del(ctx, posKey).Err()
}
