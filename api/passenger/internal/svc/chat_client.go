package svc

import (
	"context"

	"XiaoLong-Ridy/api/passenger/internal/types"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"

	"google.golang.org/grpc"
)

// ChatClient 定义 passenger API 调用 chatsvc 的最小契约（统一方案 V1.0）。
type ChatClient interface {
	// GetOrCreateConversation 获取/懒创建会话，callerID 为已鉴权的乘客 ID。
	GetOrCreateConversation(ctx context.Context, orderID int64, callerID uint64) (*types.Conversation, error)
	// ListMessages 游标分页拉取历史消息。
	ListMessages(ctx context.Context, conversationID, cursor int64, limit int32, callerID uint64) (*types.ListMessagesResponse, error)
	// SendMessage 发送消息，senderType/senderID 由网关按 JWT 注入。
	SendMessage(ctx context.Context, conversationID int64, senderType, msgType int32, senderID uint64, content, clientMsgID string) (*types.SendMessageResponse, error)
	// MarkRead 标记当前调用方已读。
	MarkRead(ctx context.Context, conversationID int64, callerID uint64) (*types.MarkReadResponse, error)
}

// GRPCChatClient 基于 chatsvc gRPC 的 ChatClient 实现。
type GRPCChatClient struct {
	client chatproto.ChatClient
}

// NewGRPCChatClient 创建 chatsvc gRPC 客户端。
func NewGRPCChatClient(conn *grpc.ClientConn) *GRPCChatClient {
	return &GRPCChatClient{client: chatproto.NewChatClient(conn)}
}

func (c *GRPCChatClient) GetOrCreateConversation(ctx context.Context, orderID int64, callerID uint64) (*types.Conversation, error) {
	resp, err := c.client.GetOrCreateConversation(ctx, &chatproto.GetOrCreateConversationRequest{
		OrderId:  orderID,
		CallerId: int64(callerID),
	})
	if err != nil {
		return nil, err
	}
	conv := &types.Conversation{
		ConversationID: resp.ConversationId,
		OrderID:        resp.OrderId,
		Status:         resp.Status,
		Opened:         resp.Opened,
		LastMsg:        resp.LastMsg,
		LastMsgAt:      resp.LastMsgAt,
		Unread:         resp.Unread,
		SenderType:     resp.SenderType,
	}
	if resp.Peer != nil {
		conv.Peer = types.PeerInfo{ID: resp.Peer.Id, Name: resp.Peer.Name}
	}
	return conv, nil
}

func (c *GRPCChatClient) ListMessages(ctx context.Context, conversationID, cursor int64, limit int32, callerID uint64) (*types.ListMessagesResponse, error) {
	resp, err := c.client.ListMessages(ctx, &chatproto.ListMessagesRequest{
		ConversationId: conversationID,
		Cursor:         cursor,
		Limit:          limit,
		CallerId:       int64(callerID),
	})
	if err != nil {
		return nil, err
	}
	out := &types.ListMessagesResponse{NextCursor: resp.NextCursor, HasMore: resp.HasMore}
	out.Messages = make([]*types.ChatMessage, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		out.Messages = append(out.Messages, &types.ChatMessage{
			ID:             m.Id,
			ConversationID: m.ConversationId,
			OrderID:        m.OrderId,
			SenderType:     m.SenderType,
			SenderID:       m.SenderId,
			MsgType:        m.MsgType,
			Content:        m.Content,
			ClientMsgID:    m.ClientMsgId,
			CreateAt:       m.CreateAt,
		})
	}
	return out, nil
}

func (c *GRPCChatClient) SendMessage(ctx context.Context, conversationID int64, senderType, msgType int32, senderID uint64, content, clientMsgID string) (*types.SendMessageResponse, error) {
	resp, err := c.client.SendMessage(ctx, &chatproto.SendMessageRequest{
		ConversationId: conversationID,
		SenderType:     senderType,
		SenderId:       int64(senderID),
		MsgType:        msgType,
		Content:        content,
		ClientMsgId:    clientMsgID,
	})
	if err != nil {
		return nil, err
	}
	return &types.SendMessageResponse{MessageID: resp.MessageId, CreateAt: resp.CreateAt}, nil
}

func (c *GRPCChatClient) MarkRead(ctx context.Context, conversationID int64, callerID uint64) (*types.MarkReadResponse, error) {
	resp, err := c.client.MarkRead(ctx, &chatproto.MarkReadRequest{
		ConversationId: conversationID,
		CallerId:       int64(callerID),
	})
	if err != nil {
		return nil, err
	}
	return &types.MarkReadResponse{Ok: resp.Ok}, nil
}
