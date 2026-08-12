package constants

// 订单状态
const (
	OrderStatusPending   = 0 // 待接单
	OrderStatusAccepted  = 1 // 已接单
	OrderStatusArrived   = 2 // 司机已到达
	OrderStatusRunning   = 3 // 行程中
	OrderStatusFinished  = 4 // 待支付
	OrderStatusPaid      = 5 // 已完成
	OrderStatusCancelled = 6 // 已取消
)

// 司机状态
const (
	DriverStatusOffline = 0 // 离线
	DriverStatusOnline  = 1 // 在线听单
	DriverStatusBusy    = 2 // 服务中
)

// 支付状态
const (
	PayStatusUnpaid = 0 // 未支付
	PayStatusPaying = 1 // 支付中
	PayStatusPaid   = 2 // 已支付
	PayStatusRefund = 3 // 已退款
)

// Redis Key 前缀
const (
	RedisKeyDriverLocation = "driver:location:%d"  // 司机实时位置
	RedisKeyOrderInfo      = "order:info:%d"       // 订单缓存
	RedisKeyUserToken      = "user:token:%s"       // 用户Token
	RedisKeyDriverNearby   = "driver:nearby:%s"    // 附近司机列表
	RedisKeySmsLimit       = "sms:limit:%s"        // 短信频率限制
)

// Kafka Topic
const (
	TopicLocationReport = "location-report" // 司机位置上报
	TopicOrderEvent     = "order-event"     // 订单事件
	TopicPushMessage    = "push-message"    // 推送消息
)
