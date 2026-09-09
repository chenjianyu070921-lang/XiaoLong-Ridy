package logic

import (
	"context"

	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	driverproto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

// ChatLogic 封装乘客端聊天业务流程：会话获取、历史消息、发送、已读、快捷短语。
// 统一方案 V1.0：乘客端 MVP 采用轮询（非 WebSocket），底层由 rpc/chatsvc 提供能力。
type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
	logx.Logger
}

// NewChatLogic 创建聊天逻辑实例。token 由 handler 从 Bearer 头提取后注入。
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *ChatLogic {
	return &ChatLogic{ctx: ctx, svcCtx: svcCtx, token: token, Logger: logx.WithContext(ctx)}
}

func (l *ChatLogic) userID() (uint64, error) {
	uid, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

// GetConversation 获取/懒创建会话，并补全对端（司机）昵称。
func (l *ChatLogic) GetConversation(orderID int64) (*types.Conversation, error) {
	uid, err := l.userID()
	if err != nil {
		return nil, err
	}
	if l.svcCtx.ChatClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	conv, err := l.svcCtx.ChatClient.GetOrCreateConversation(l.ctx, orderID, uid)
	if err != nil {
		return nil, err
	}
	if conv.Peer.ID > 0 && l.svcCtx.DriverClient != nil {
		if d, derr := l.svcCtx.DriverClient.GetDriver(l.ctx, &driverproto.GetDriverRequest{Id: conv.Peer.ID}); derr == nil && d != nil && d.GetDriver() != nil {
			conv.Peer.Name = d.GetDriver().GetRealName()
		}
	}
	return conv, nil
}

// ListMessages 游标分页拉取历史消息。
func (l *ChatLogic) ListMessages(conversationID, cursor int64, limit int32) (*types.ListMessagesResponse, error) {
	uid, err := l.userID()
	if err != nil {
		return nil, err
	}
	if l.svcCtx.ChatClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	return l.svcCtx.ChatClient.ListMessages(l.ctx, conversationID, cursor, limit, uid)
}

// SendMessage 发送消息，senderType/senderID 由网关按 JWT 注入（乘客=2）。
func (l *ChatLogic) SendMessage(conversationID int64, msgType int32, content, clientMsgID string) (*types.SendMessageResponse, error) {
	uid, err := l.userID()
	if err != nil {
		return nil, err
	}
	if l.svcCtx.ChatClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.ChatClient.SendMessage(l.ctx, conversationID, types.ChatSenderPassenger, msgType, uid, content, clientMsgID)
}

// MarkRead 标记当前乘客已读。
func (l *ChatLogic) MarkRead(conversationID int64) (*types.MarkReadResponse, error) {
	uid, err := l.userID()
	if err != nil {
		return nil, err
	}
	if l.svcCtx.ChatClient == nil {
		return nil, ErrOrderClientNotConfigured
	}
	return l.svcCtx.ChatClient.MarkRead(l.ctx, conversationID, uid)
}

// QuickPhrases 返回 MVP 硬编码的 5 条快捷短语。
func (l *ChatLogic) QuickPhrases() *types.ChatQuickPhrasesResponse {
	return &types.ChatQuickPhrasesResponse{Phrases: []types.ChatQuickPhrase{
		{ID: 1, Content: "我马上到"},
		{ID: 2, Content: "我在附近"},
		{ID: 3, Content: "请稍等"},
		{ID: 4, Content: "我已到达上车点"},
		{ID: 5, Content: "请打开车门"},
	}}
}
