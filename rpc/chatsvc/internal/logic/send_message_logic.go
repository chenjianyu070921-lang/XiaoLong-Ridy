package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/sensitive"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// SendMessageLogic 处理消息发送：敏感词拦截 + 幂等 + 未读累加。
type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSendMessageLogic 创建逻辑实例。
func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// SendMessage 发送一条消息。命中敏感词返回 InvalidArgument（网关映射为 46000），
// 重复 client_msg_id 幂等返回首次结果。
func (l *SendMessageLogic) SendMessage(in *chatproto.SendMessageRequest) (*chatproto.SendMessageResponse, error) {
	if in.ConversationId <= 0 || in.SenderId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid params")
	}
	if strings.TrimSpace(in.Content) == "" || strings.TrimSpace(in.ClientMsgId) == "" {
		return nil, status.Error(codes.InvalidArgument, "content and client_msg_id required")
	}
	if in.MsgType != MsgTypeText && in.MsgType != MsgTypeQuick {
		return nil, status.Error(codes.InvalidArgument, "unsupported msg_type")
	}
	if sensitive.Contains(in.Content) {
		return nil, status.Error(codes.InvalidArgument, "命中敏感词，消息已拦截")
	}

	var conv model.Conversation
	if err := l.svcCtx.DB.First(&conv, "id = ?", in.ConversationId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "conversation not found")
		}
		return nil, status.Error(codes.Internal, "query conversation failed")
	}

	senderType := SenderTypeDriver
	switch {
	case int64(conv.PassengerId) == in.SenderId:
		senderType = SenderTypePassenger
	case int64(conv.DriverId) == in.SenderId:
		senderType = SenderTypeDriver
	default:
		return nil, status.Error(codes.PermissionDenied, "forbidden: not participant")
	}
	if conv.Status == int8(ConvStatusArchived) || conv.Status == int8(ConvStatusClosed) {
		return nil, status.Error(codes.FailedPrecondition, "会话已结束，不可发送")
	}

	// 幂等：client_msg_id 重复则直接返回首次结果。
	var existing model.Message
	if err := l.svcCtx.DB.Where("client_msg_id = ?", in.ClientMsgId).First(&existing).Error; err == nil {
		return &chatproto.SendMessageResponse{MessageId: int64(existing.Id), CreateAt: existing.CreateAt.Unix()}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.Internal, "query message failed")
	}

	now := time.Now()
	msg := model.Message{
		ConversationId: int64(conv.Id),
		OrderId:        conv.OrderId,
		SenderType:     int8(senderType),
		SenderId:       in.SenderId,
		MsgType:        int8(in.MsgType),
		Content:        in.Content,
		ClientMsgId:    in.ClientMsgId,
		CreateAt:       now,
	}
	if err := l.svcCtx.DB.Create(&msg).Error; err != nil {
		// 并发插入触发唯一键冲突时回查首次结果。
		var dup model.Message
		if e2 := l.svcCtx.DB.Where("client_msg_id = ?", in.ClientMsgId).First(&dup).Error; e2 == nil {
			return &chatproto.SendMessageResponse{MessageId: int64(dup.Id), CreateAt: dup.CreateAt.Unix()}, nil
		}
		return nil, status.Error(codes.Internal, "insert message failed")
	}

	updates := map[string]any{"last_msg": in.Content, "last_msg_at": now}
	if senderType == SenderTypeDriver {
		updates["unread_passenger"] = gorm.Expr("unread_passenger + 1")
	} else {
		updates["unread_driver"] = gorm.Expr("unread_driver + 1")
	}
	l.svcCtx.DB.Model(&conv).Updates(updates)

	return &chatproto.SendMessageResponse{MessageId: int64(msg.Id), CreateAt: now.Unix()}, nil
}
