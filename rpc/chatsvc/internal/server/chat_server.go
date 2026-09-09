package server

import (
	"context"

	"XiaoLong-Ridy/rpc/chatsvc/internal/logic"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"
)

// ChatServer 实现 chatsvc.proto 定义的 Chat 服务。
type ChatServer struct {
	svcCtx *svc.ServiceContext
	chatproto.UnimplementedChatServer
}

// NewChatServer 创建 Chat 服务实例。
func NewChatServer(svcCtx *svc.ServiceContext) *ChatServer {
	return &ChatServer{svcCtx: svcCtx}
}

func (s *ChatServer) GetOrCreateConversation(ctx context.Context, in *chatproto.GetOrCreateConversationRequest) (*chatproto.GetOrCreateConversationResponse, error) {
	return logic.NewGetOrCreateConversationLogic(ctx, s.svcCtx).GetOrCreateConversation(in)
}

func (s *ChatServer) ListMessages(ctx context.Context, in *chatproto.ListMessagesRequest) (*chatproto.ListMessagesResponse, error) {
	return logic.NewListMessagesLogic(ctx, s.svcCtx).ListMessages(in)
}

func (s *ChatServer) SendMessage(ctx context.Context, in *chatproto.SendMessageRequest) (*chatproto.SendMessageResponse, error) {
	return logic.NewSendMessageLogic(ctx, s.svcCtx).SendMessage(in)
}

func (s *ChatServer) MarkRead(ctx context.Context, in *chatproto.MarkReadRequest) (*chatproto.MarkReadResponse, error) {
	return logic.NewMarkReadLogic(ctx, s.svcCtx).MarkRead(in)
}
