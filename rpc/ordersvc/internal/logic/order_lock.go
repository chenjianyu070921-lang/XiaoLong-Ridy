package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/common/constants"

	"github.com/redis/go-redis/v9"
)

func acquireOrderLock(ctx context.Context, rdb *redis.Client, orderID uint64) (func(), error) {
	if rdb == nil {
		return func() {}, nil
	}
	key := fmt.Sprintf(constants.RedisOrderLock, orderID)
	ok, err := rdb.SetNX(ctx, key, "1", 10*time.Second).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrOrderStatusNotAllowed
	}
	return func() { _ = rdb.Del(context.Background(), key).Err() }, nil
}
