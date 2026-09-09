package keyutil

import (
	"fmt"
	"sync/atomic"
	"time"
)

var orderSequence uint64

// GenOrderID 生成订单号，格式为 DD + 秒级时间 + 纳秒 + 递增序列，长度不超过 ride_order.order_no 的 32 位限制。
func GenOrderID() string {
	now := time.Now()
	seq := atomic.AddUint64(&orderSequence, 1) % 1000000
	return fmt.Sprintf("DD%s%09d%06d", now.Format("20060102150405"), now.Nanosecond(), seq)
}
