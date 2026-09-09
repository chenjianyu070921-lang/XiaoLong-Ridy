package logic

import (
	"context"

	__proto "XiaoLong-Ridy/rpc/chatsvc/proto"
	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListMessages 游标分页拉取消息（按 id 升序）。caller 字段用于权限校验。
func (l *ListMessagesLogic) ListMessages(in *__proto.ListMessagesRequest) (*__proto.ListMessagesResponse, error) {
	var conv model.IMConversation
	if err := l.svcCtx.DB.Where("id = ?", in.ConversationId).First(&conv).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "会话不存在")
	}
	isDriver := in.CallerType == 1
	isPassenger := in.CallerType == 2
	if (isDriver && in.CallerId != conv.DriverId) || (isPassenger && in.CallerId != conv.PassengerId) {
		return nil, status.Errorf(codes.PermissionDenied, "无权限读取该会话消息")
	}

	limit := int(in.Limit)
	if limit <= 0 || limit > 20 {
		limit = 20
	}

	var msgs []model.IMMessage
	q := l.svcCtx.DB.Where("conversation_id = ?", in.ConversationId)
	if in.Cursor > 0 {
		q = q.Where("id > ?", in.Cursor)
	}
	if err := q.Order("id asc").Limit(limit + 1).Find(&msgs).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询消息失败: %v", err)
	}

	hasMore := false
	if len(msgs) > limit {
		hasMore = true
		msgs = msgs[:limit]
	}

	resp := &__proto.ListMessagesResponse{HasMore: hasMore}
	if len(msgs) > 0 {
		resp.NextCursor = msgs[len(msgs)-1].Id
	}
	for _, m := range msgs {
		resp.Messages = append(resp.Messages, &__proto.ChatMessage{
			Id:             m.Id,
			ConversationId: m.ConversationId,
			OrderId:        m.OrderId,
			SenderType:     int32(m.SenderType),
			SenderId:       m.SenderId,
			MsgType:        int32(m.MsgType),
			Content:        m.Content,
			ClientMsgId:     m.ClientMsgId,
			CreateAt:       m.CreatedAt.Unix(),
		})
	}
	return resp, nil
}
