package logic

import "XiaoLong-Ridy/common/constants"

// CanTransit 集中校验订单状态流转，避免各业务入口分散维护状态规则。
func CanTransit(from, to int8) bool {
	switch from {
	case 0:
		return to == constants.OrderStatusWaitAccept
	case constants.OrderStatusWaitAccept:
		return to == constants.OrderStatusAccepted || to == constants.OrderStatusCancelled
	case constants.OrderStatusAccepted:
		return to == constants.OrderStatusOnTrip || to == constants.OrderStatusCancelled
	case constants.OrderStatusOnTrip:
		return to == constants.OrderStatusWaitPay
	case constants.OrderStatusWaitPay:
		return to == constants.OrderStatusCompleted
	case constants.OrderStatusCompleted, constants.OrderStatusCancelled:
		return false
	default:
		return false
	}
}
