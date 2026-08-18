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

// 派单类型
const (
	DispatchTypeAuto = 1 // 自动派单
)

// 派单记录状态
const (
	DispatchStatusPending   = 1 // 派单中
	DispatchStatusAccepted  = 2 // 已接受
	DispatchStatusCancelled = 5 // 已取消
)

// 订单操作方
const (
	OperatorUser   = "user"
	OperatorDriver = "driver"
	OperatorSystem = "system"
	OperatorAdmin  = "admin"
	RedisDriverPos    = "driver:pos:%d"   // 司机位置
	RedisOrderInfo    = "order:info:%d"   // 订单信息
	RedisSmsCode      = "sms:code:%s"     // 验证码
	RedisOrderLock    = "r:lock:order:%d" // 接单分布式锁
	RedisDriverGeo    = "driver:geo:%s"   // 司机位置 GEO，按城市
	RedisDriverOnline = "driver:online"   // 在线司机集合
)

// Kafka/事件主题
const (
	TopicLocation           = "location-report"      // 司机位置上报
	TopicOrder              = "order-event"          // 订单事件
	TopicOrderCreated       = "order.created"        // 订单创建
	TopicOrderStatusChanged = "order.status.changed" // 订单状态变更
	TopicOrderCancelled     = "order.canceled"       // 订单取消
	TopicDispatchNew        = "dispatch.new"         // 派单通知
	TopicDispatchResult     = "dispatch.result"      // 派单结果
)

// 事件流
const (
	OrderEventStream     = "order:event:stream"     // 订单事件流
	DriverLocationStream = "driver:location:stream" // 司机位置流
)
