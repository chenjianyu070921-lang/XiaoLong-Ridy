package keyutil

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GenerateOrderID 生成订单号
func GenerateOrderID() string {
	now := time.Now()
	return fmt.Sprintf("%s%04d%03d",
		now.Format("20060102150405"),
		rand.Intn(10000),
		rand.Intn(1000),
	)
}

// GenerateDriverID 生成司机ID
func GenerateDriverID() string {
	return fmt.Sprintf("D%013d", time.Now().UnixNano()%10000000000000+rand.Int63n(100000))
}

// GenerateUserID 生成用户ID
func GenerateUserID() string {
	return fmt.Sprintf("U%013d", time.Now().UnixNano()%10000000000000+rand.Int63n(100000))
}
