package logic

import (
	"context"
	"time"

	__proto "XiaoLong-Ridy/rpc/chatsvc/proto"
	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetOrCreateConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrCreateConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateConversationLogic {
	return &GetOrCreateConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrCreateConversation 按订单懒创建/获取会话。
// 权限：仅订单的司机或乘客可访问；状态映射（图片微调后）：
//   ACCEPTED/ON_TRIP/WAIT_PAY → 进行中(1)；COMPLETED → 已归档(2)；其余 → 非 opened。
func (l *GetOrCreateConversationLogic) GetOrCreateConversation(in *__proto.GetOrCreateConversationRequest) (*__proto.GetOrCreateConversationResponse, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 3*time.Second)
	defer cancel()

	order, err := l.svcCtx.OrderClient.GetOrder(ctx, &orderproto.GetOrderRequest{OrderId: in.OrderId})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "查询订单失败: %v", err)
	}

	isDriver := in.CallerType == 1
	isPassenger := in.CallerType == 2
	if (isDriver && in.CallerId != order.DriverId) || (isPassenger && in.CallerId != order.UserId) {
		return nil, status.Errorf(codes.PermissionDenied, "无权限访问该订单会话")
	}

	convStatus, opened := mapOrderStatusToConversation(order.Status)
	if !opened {
		return &__proto.GetOrCreateConversationResponse{
			OrderId: in.OrderId,
			OrderNo: order.OrderNo,
			Status:  int32(convStatus),
			Opened:  false,
		}, nil
	}

	var conv model.IMConversation
	err = l.svcCtx.DB.
		Where("order_id = ?", in.OrderId).
		Attrs(model.IMConversation{
			OrderId:     in.OrderId,
			OrderNo:     order.OrderNo,
			DriverId:    order.DriverId,
			PassengerId: order.UserId,
			Status:      int8(convStatus),
		}).
		FirstOrCreate(&conv).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建会话失败: %v", err)
	}

	// 订单状态推进后同步会话状态（如行程完成→归档只读）。
	if int8(convStatus) != conv.Status {
		_ = l.svcCtx.DB.Model(&model.IMConversation{}).Where("id = ?", conv.Id).Update("status", int8(convStatus))
		conv.Status = int8(convStatus)
	}

	var unread int32
	var peerId int64
	peerRole := "passenger"
	peerName := "乘客"
	if isDriver {
		unread = int32(conv.UnreadDriver)
		peerId = order.UserId
	} else {
		unread = int32(conv.UnreadPassenger)
		peerId = order.DriverId
		peerRole = "driver"
		peerName = "司机"
	}

	var lastMsgAt int64
	if conv.LastMsgAt != nil {
		lastMsgAt = conv.LastMsgAt.Unix()
	}

	return &__proto.GetOrCreateConversationResponse{
		ConversationId: conv.Id,
		OrderId:        conv.OrderId,
		OrderNo:        conv.OrderNo,
		Status:         int32(conv.Status),
		LastMsg:        conv.LastMsg,
		LastMsgAt:      lastMsgAt,
		Unread:         unread,
		Peer:          &__proto.PeerInfo{Id: peerId, Name: peerName, Role: peerRole},
		Opened:         true,
	}, nil
}

// mapOrderStatusToConversation 将订单状态映射为会话状态；返回 (会话状态, 是否可聊天)。
func mapOrderStatusToConversation(s orderproto.OrderStatus) (int, bool) {
	switch s {
	case orderproto.OrderStatus_ORDER_STATUS_ACCEPTED,
		orderproto.OrderStatus_ORDER_STATUS_ON_TRIP,
		orderproto.OrderStatus_ORDER_STATUS_WAIT_PAY:
		return 1, true
	case orderproto.OrderStatus_ORDER_STATUS_COMPLETED:
		return 2, true
	case orderproto.OrderStatus_ORDER_STATUS_CANCELLED,
		orderproto.OrderStatus_ORDER_STATUS_REFUNDED:
		return 3, false
	default: // WAIT_ACCEPT / UNSPECIFIED
		return 1, false
	}
}
