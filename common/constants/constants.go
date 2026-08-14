package constants

// 订单状态
const (
	OrderWait   = 0 // 等待接单
	OrderAccept = 1 // 已接单
	OrderGoing  = 2 // 行程中
	OrderDone   = 3 // 已完成
	OrderCancel = 4 // 已取消
)

// Redis Key 模板（用的时候 fmt.Sprintf 填入 ID）
const (
	RedisDriverPos  = "driver:pos:%d"  // 司机位置
	RedisOrderInfo  = "order:info:%d"  // 订单信息
	RedisSmsCode    = "sms:code:%s"    // 验证码
)

// Kafka 消息主题
const (
	TopicLocation   = "location-report" // 司机位置上报
	TopicOrder      = "order-event"     // 订单事件
	TopicOrderPaid  = "order.paid"      // 订单支付成功事件
)
