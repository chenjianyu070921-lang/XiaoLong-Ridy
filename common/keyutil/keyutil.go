package keyutil

import (
	"fmt"
	"math/rand"
	"time"
)

// GenOrderID 生成订单号 例: 202408121530051234
func GenOrderID() string {
	return fmt.Sprintf("%s%04d", time.Now().Format("20060102150405"), rand.Intn(10000))
}
