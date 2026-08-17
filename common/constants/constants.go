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
)
