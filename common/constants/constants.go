package constants

// 订单状态
const (
	OrderStatusWaitAccept = 1 // 待接单
	OrderStatusAccepted   = 2 // 已接单
	OrderStatusOnTrip     = 3 // 行程中
	OrderStatusWaitPay    = 4 // 待支付
	OrderStatusCompleted  = 5 // 已完成
	OrderStatusCancelled  = 6 // 已取消
)

// 订单操作方
const (
	OperatorUser   = "user"
	OperatorDriver = "driver"
	OperatorSystem = "system"
	OperatorAdmin  = "admin"
)

// Redis Key 模板（用的时候 fmt.Sprintf 填入 ID）
const (
	RedisDriverPos = "driver:pos:%d" // 司机位置
	RedisOrderInfo = "order:info:%d" // 订单信息
	RedisSmsCode   = "sms:code:%s"   // 验证码
)

// Kafka 消息主题
const (
	TopicLocation   = "location-report" // 司机位置上报
	TopicOrder      = "order-event"     // 订单事件
	TopicOrderPaid  = "order.paid"      // 订单支付成功事件
)
