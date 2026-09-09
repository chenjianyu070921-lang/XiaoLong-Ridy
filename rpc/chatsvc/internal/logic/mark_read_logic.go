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

// MarkReadLogic 处理标记已读：清零调用方未读数。
type MarkReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewMarkReadLogic 创建逻辑实例。
func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// MarkRead 标记当前调用方已读。
func (l *MarkReadLogic) MarkRead(in *chatproto.MarkReadRequest) (*chatproto.MarkReadResponse, error) {
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
	if int64(conv.DriverId) == in.CallerId {
		l.svcCtx.DB.Model(&conv).Update("unread_driver", 0)
	} else {
		l.svcCtx.DB.Model(&conv).Update("unread_passenger", 0)
	}
	return &chatproto.MarkReadResponse{Ok: true}, nil
}
