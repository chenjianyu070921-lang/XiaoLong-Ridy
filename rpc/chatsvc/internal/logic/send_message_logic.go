package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"XiaoLong-Ridy/common/constants"
	__proto "XiaoLong-Ridy/rpc/chatsvc/proto"
	"XiaoLong-Ridy/rpc/chatsvc/internal/model"
	"XiaoLong-Ridy/rpc/chatsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// chatPushPayload 是经 Redis Pub/Sub 推送给司机 WS（driver:push:%d）的负载。
type chatPushPayload struct {
	Type           string          `json:"type"` // 固定 "chat.message"
	ConversationId int64           `json:"conversationId"`
	OrderId        int64           `json:"orderId"`
	OrderNo        string          `json:"orderNo"`
	Message        chatPushMessage `json:"message"`
}

type chatPushMessage struct {
	Id          int64  `json:"id"`
	SenderType  int32  `json:"senderType"`
	SenderId    int64  `json:"senderId"`
	MsgType     int32  `json:"msgType"`
	Content     string `json:"content"`
	ClientMsgId string `json:"clientMsgId"`
	CreateAt    int64  `json:"createAt"`
}

// SendMessage 发送消息：权限校验 → 状态闸门 → 敏感词拦截 → 幂等(client_msg_id) → 落库 → 实时推送。
func (l *SendMessageLogic) SendMessage(in *__proto.SendMessageRequest) (*__proto.SendMessageResponse, error) {
	var conv model.IMConversation
	if err := l.svcCtx.DB.Where("id = ?", in.ConversationId).First(&conv).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "会话不存在")
	}
	if conv.Status != 1 {
		return nil, status.Errorf(codes.FailedPrecondition, "当前会话不可发送消息（已归档或已关闭）")
	}

	isDriver := in.SenderType == 1
	if (isDriver && in.SenderId != conv.DriverId) || (!isDriver && in.SenderId != conv.PassengerId) {
		return nil, status.Errorf(codes.PermissionDenied, "无权限向该会话发送消息")
	}

	content := strings.TrimSpace(in.Content)
	if content == "" {
		return nil, status.Errorf(codes.InvalidArgument, "消息内容不能为空")
	}
	if len([]rune(content)) > 500 {
		content = string([]rune(content)[:500])
	}
	if l.svcCtx.ContainsSensitive(content) {
		return nil, status.Errorf(codes.InvalidArgument, "消息包含敏感词，禁止线下交易或交换联系方式")
	}

	// 幂等：同一 client_msg_id 只落一条，重复提交返回首次结果。
	var existing model.IMMessage
	if err := l.svcCtx.DB.Where("client_msg_id = ?", in.ClientMsgId).First(&existing).Error; err == nil {
		return &__proto.SendMessageResponse{MessageId: existing.Id, CreateAt: existing.CreatedAt.Unix()}, nil
	}

	now := time.Now()
	msg := model.IMMessage{
		ConversationId: conv.Id,
		OrderId:        conv.OrderId,
		SenderType:     int8(in.SenderType),
		SenderId:       in.SenderId,
		MsgType:        int8(in.MsgType),
		Content:        content,
		ClientMsgId:    in.ClientMsgId,
		CreatedAt:      now,
	}
	if err := l.svcCtx.DB.Create(&msg).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "保存消息失败: %v", err)
	}

	update := map[string]interface{}{
		"last_msg":    content,
		"last_msg_at": now,
	}
	if isDriver {
		update["unread_passenger"] = gorm.Expr("unread_passenger + 1")
	} else {
		update["unread_driver"] = gorm.Expr("unread_driver + 1")
	}
	_ = l.svcCtx.DB.Model(&model.IMConversation{}).Where("id = ?", conv.Id).Updates(update)

	// 实时转发：发布到司机端 WS 频道 driver:push:{driverId}。
	payload := chatPushPayload{
		Type:           "chat.message",
		ConversationId: conv.Id,
		OrderId:        conv.OrderId,
		OrderNo:        conv.OrderNo,
		Message: chatPushMessage{
			Id:          msg.Id,
			SenderType:  int32(msg.SenderType),
			SenderId:    msg.SenderId,
			MsgType:     int32(msg.MsgType),
			Content:     msg.Content,
			ClientMsgId: msg.ClientMsgId,
			CreateAt:    msg.CreatedAt.Unix(),
		},
	}
	if data, e := json.Marshal(payload); e == nil {
		l.svcCtx.RedisClient.Publish(context.Background(), fmt.Sprintf(constants.RedisDriverPush, conv.DriverId), string(data))
	}

	return &__proto.SendMessageResponse{MessageId: msg.Id, CreateAt: now.Unix()}, nil
}
