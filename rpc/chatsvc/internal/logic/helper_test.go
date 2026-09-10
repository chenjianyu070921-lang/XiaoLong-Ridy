package logic

import (
	"testing"

	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"
)

// TestIsConversationTerminal 验证单一终态判定的唯一真相来源。
func TestIsConversationTerminal(t *testing.T) {
	cases := []struct {
		status int32
		want   bool
	}{
		{ConvStatusOngoing, false},
		{ConvStatusArchived, true},
		{ConvStatusClosed, true},
		{99, false}, // 未知态默认非终态，避免误杀可聊天会话
	}
	for _, c := range cases {
		if got := IsConversationTerminal(c.status); got != c.want {
			t.Errorf("IsConversationTerminal(%d) = %v, want %v", c.status, got, c.want)
		}
	}
}

// TestOrderView 验证 orderView 的角色解析、参与方判定与终态推导（纯逻辑，无外部依赖）。
func TestOrderView(t *testing.T) {
	mkOrder := func(driverId, userId int64, st orderproto.OrderStatus) *orderproto.GetOrderResponse {
		return &orderproto.GetOrderResponse{DriverId: driverId, UserId: userId, Status: st}
	}
	cases := []struct {
		name            string
		order           *orderproto.GetOrderResponse
		callerID        int64
		wantConv        int32
		wantOpened      bool
		wantReadonly    bool
		wantSender      int32
		wantPeer        int32
		wantParticipant bool
	}{
		{
			name:            "司机视角-已接单",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_ACCEPTED),
			callerID:        8,
			wantConv:        ConvStatusOngoing, wantOpened: true, wantReadonly: false,
			wantSender: SenderTypeDriver, wantPeer: 20, wantParticipant: true,
		},
		{
			name:            "乘客视角-已接单",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_ACCEPTED),
			callerID:        20,
			wantConv:        ConvStatusOngoing, wantOpened: true, wantReadonly: false,
			wantSender: SenderTypePassenger, wantPeer: 8, wantParticipant: true,
		},
		{
			name:            "司机视角-待接单（未开启但参与）",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT),
			callerID:        8,
			wantConv:        ConvStatusOngoing, wantOpened: false, wantReadonly: false,
			wantSender: SenderTypeDriver, wantPeer: 20, wantParticipant: true,
		},
		{
			name:            "非参与方",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_ACCEPTED),
			callerID:        999,
			wantConv:        ConvStatusOngoing, wantOpened: true, wantReadonly: false,
			wantSender: SenderTypePassenger, wantPeer: 8, wantParticipant: false,
		},
		{
			name:            "司机视角-已完成（终态归档）",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_COMPLETED),
			callerID:        8,
			wantConv:        ConvStatusArchived, wantOpened: true, wantReadonly: true,
			wantSender: SenderTypeDriver, wantPeer: 20, wantParticipant: true,
		},
		{
			name:            "司机视角-已取消（终态关闭）",
			order:           mkOrder(8, 20, orderproto.OrderStatus_ORDER_STATUS_CANCELLED),
			callerID:        8,
			wantConv:        ConvStatusClosed, wantOpened: false, wantReadonly: true,
			wantSender: SenderTypeDriver, wantPeer: 20, wantParticipant: true,
		},
		{
			name:            "nil 订单",
			order:           nil,
			callerID:        8,
			wantConv:        ConvStatusOngoing, wantOpened: false, wantReadonly: false,
			wantSender: 0, wantPeer: 0, wantParticipant: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			convStatus, opened, readonly, senderType, peerID, isParticipant := orderView(c.order, c.callerID)
			if convStatus != c.wantConv || opened != c.wantOpened || readonly != c.wantReadonly ||
				senderType != c.wantSender || peerID != c.wantPeer || isParticipant != c.wantParticipant {
				t.Errorf("orderView = (%d,%v,%v,%d,%d,%v), want (%d,%v,%v,%d,%d,%v)",
					convStatus, opened, readonly, senderType, peerID, isParticipant,
					c.wantConv, c.wantOpened, c.wantReadonly, c.wantSender, c.wantPeer, c.wantParticipant)
			}
			// 终态一致性：readonly 必须与 IsConversationTerminal(convStatus) 同义
			if readonly != IsConversationTerminal(convStatus) {
				t.Errorf("readonly(%v) 与 IsConversationTerminal(%d)=%v 不一致",
					readonly, convStatus, IsConversationTerminal(convStatus))
			}
		})
	}
}
