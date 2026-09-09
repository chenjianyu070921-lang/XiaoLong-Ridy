package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ListMessagesLogic 处理历史消息游标分页。
type ListMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewListMessagesLogic 创建逻辑实例。
func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ListMessages 按 id 倒序取一页再翻转为升序；cursor 为上一页最小 id，0 表示最新一页。
func (l *ListMessagesLogic) ListMessages(in *chatproto.ListMessagesRequest) (*chatproto.ListMessagesResponse, error) {
	if in.ConversationId <= 0 || in.CallerId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid params")
	}
	var conv model.Conversation
	if err := l.svcCtx.DB.First(&conv, "id = ?", in.ConversationId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "conversation not found")
		}
		return nil, status.Error(codes.Internal, "query conversation failed")
	}
	if int64(conv.DriverId) != in.CallerId && int64(conv.PassengerId) != in.CallerId {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}

	limit := in.Limit
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	q := l.svcCtx.DB.Where("conversation_id = ?", conv.Id).Order("id DESC")
	if in.Cursor > 0 {
		q = q.Where("id < ?", in.Cursor)
	}
	var rows []model.Message
	if err := q.Limit(int(limit)).Find(&rows).Error; err != nil {
		return nil, status.Error(codes.Internal, "query messages failed")
	}

	// 倒序取回，翻转为升序返回。
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	msgs := make([]*chatproto.ChatMessage, 0, len(rows))
	var nextCursor int64
	for _, m := range rows {
		msgs = append(msgs, toChatMessage(m))
		nextCursor = int64(m.Id)
	}
	hasMore := false
	if len(rows) == int(limit) {
		var cnt int64
		if l.svcCtx.DB.Model(&model.Message{}).
			Where("conversation_id = ? AND id < ?", conv.Id, nextCursor).
			Count(&cnt); cnt > 0 {
			hasMore = true
		}
	}
	return &chatproto.ListMessagesResponse{Messages: msgs, NextCursor: nextCursor, HasMore: hasMore}, nil
}

func toChatMessage(m model.Message) *chatproto.ChatMessage {
	return &chatproto.ChatMessage{
		Id:             int64(m.Id),
		ConversationId: int64(m.ConversationId),
		OrderId:        m.OrderId,
		SenderType:     int32(m.SenderType),
		SenderId:       m.SenderId,
		MsgType:        int32(m.MsgType),
		Content:        m.Content,
		ClientMsgId:    m.ClientMsgId,
		CreateAt:       m.CreateAt.Unix(),
	}
}
