package logic

import (
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// 统一方案 V1.0 枚举定义（两端代码共用，禁止第二套编号）。
const (
	// SenderTypeDriver 发送方为司机。
	SenderTypeDriver int32 = 1
	// SenderTypePassenger 发送方为乘客。
	SenderTypePassenger int32 = 2

	// MsgTypeText 文本消息。
	MsgTypeText int32 = 1
	// MsgTypeQuick 快捷短语。
	MsgTypeQuick int32 = 2

	// ConvStatusOngoing 会话进行中。
	ConvStatusOngoing int32 = 1
	// ConvStatusArchived 会话已归档（订单完成，只读）。
	ConvStatusArchived int32 = 2
	// ConvStatusClosed 会话已关闭（订单取消/退款）。
	ConvStatusClosed int32 = 3
)

// orderView 从订单响应中解析出调用方在会话中的角色、对端 ID、会话状态与是否可聊天。
// callerID 为已鉴权的调用方 ID（司机或乘客）。
func orderView(order *orderproto.GetOrderResponse, callerID int64) (
	convStatus int32, opened, readonly bool, senderType, peerID int32, isParticipant bool,
) {
	if order == nil {
		return ConvStatusOngoing, false, false, 0, 0, false
	}
	isParticipant = order.UserId == callerID || order.DriverId == callerID
	if order.DriverId == callerID {
		senderType = SenderTypeDriver
		peerID = int32(order.UserId)
	} else {
		senderType = SenderTypePassenger
		peerID = int32(order.DriverId)
	}
	switch order.Status {
	case orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT:
		// 等待接单阶段：不建会话，返回未开启。
		convStatus, opened, readonly = ConvStatusOngoing, false, false
	case orderproto.OrderStatus_ORDER_STATUS_ACCEPTED,
		orderproto.OrderStatus_ORDER_STATUS_ON_TRIP,
		orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY:
		convStatus, opened, readonly = ConvStatusOngoing, true, false
	case orderproto.OrderStatus_ORDER_STATUS_COMPLETED:
		convStatus, opened, readonly = ConvStatusArchived, true, true
	default: // CANCELLED / REFUNDED 等终态
		convStatus, opened, readonly = ConvStatusClosed, false, true
	}
	return
}
