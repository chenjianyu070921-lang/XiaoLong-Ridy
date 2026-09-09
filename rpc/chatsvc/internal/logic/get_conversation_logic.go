package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetOrCreateConversationLogic 处理会话懒创建与查询。
type GetOrCreateConversationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewGetOrCreateConversationLogic 创建逻辑实例。
func NewGetOrCreateConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateConversationLogic {
	return &GetOrCreateConversationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOrCreateConversation 司机接单后首次访问时懒创建会话；等待接单阶段返回未开启。
func (l *GetOrCreateConversationLogic) GetOrCreateConversation(in *chatproto.GetOrCreateConversationRequest) (*chatproto.GetOrCreateConversationResponse, error) {
	if in.OrderId <= 0 || in.CallerId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id or caller_id")
	}
	if l.svcCtx.OrderClient == nil {
		return nil, status.Error(codes.Unavailable, "order service unavailable")
	}
	order, err := l.svcCtx.OrderClient.GetOrder(l.ctx, &orderproto.GetOrderRequest{OrderId: in.OrderId})
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	convStatus, opened, _, senderType, peerID, isParticipant := orderView(order, in.CallerId)
	if !isParticipant {
		return nil, status.Error(codes.PermissionDenied, "forbidden: not order participant")
	}

	orderIDStr := strconv.FormatInt(in.OrderId, 10)
	resp := &chatproto.GetOrCreateConversationResponse{
		OrderId:    orderIDStr,
		Status:     convStatus,
		Opened:     opened,
		SenderType: senderType,
		Peer:       &chatproto.PeerInfo{Id: int64(peerID)},
	}

	if !opened {
		// 等待接单 / 已关闭：不建会话，直接返回未开启。
		return resp, nil
	}

	// 按订单号懒创建会话（幂等）。
	// 先查询；不存在则 INSERT IGNORE，再回查，避免 FirstOrCreate 在并发下触发唯一键冲突。
	var conv model.Conversation
	if err := l.svcCtx.DB.Where("order_id = ?", orderIDStr).First(&conv).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Internal, "query conversation failed")
		}
		conv = model.Conversation{
			OrderId:     orderIDStr,
			DriverId:    int64(order.DriverId),
			PassengerId: int64(order.UserId),
			Status:      int8(ConvStatusOngoing),
			CreateAt:    time.Now(),
		}
		if createErr := l.svcCtx.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "order_id"}},
			DoNothing: true,
		}).Create(&conv).Error; createErr != nil {
			return nil, status.Error(codes.Internal, "create conversation failed")
		}
		if findErr := l.svcCtx.DB.Where("order_id = ?", orderIDStr).First(&conv).Error; findErr != nil {
			return nil, status.Error(codes.Internal, "query conversation failed")
		}
	}

	resp.ConversationId = int64(conv.Id)
	resp.Status = int32(conv.Status)
	resp.LastMsg = conv.LastMsg
	if conv.LastMsgAt != nil {
		resp.LastMsgAt = conv.LastMsgAt.Unix()
	}
	if senderType == SenderTypeDriver {
		resp.Unread = int32(conv.UnreadDriver)
	} else {
		resp.Unread = int32(conv.UnreadPassenger)
	}
	return resp, nil
}
