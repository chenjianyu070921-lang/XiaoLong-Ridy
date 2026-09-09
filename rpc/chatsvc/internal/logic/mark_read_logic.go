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

type MarkReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkRead 标记调用方未读数为 0。
func (l *MarkReadLogic) MarkRead(in *__proto.MarkReadRequest) (*__proto.MarkReadResponse, error) {
	var conv model.IMConversation
	if err := l.svcCtx.DB.Where("id = ?", in.ConversationId).First(&conv).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "会话不存在")
	}
	isDriver := in.CallerType == 1
	isPassenger := in.CallerType == 2
	if (isDriver && in.CallerId != conv.DriverId) || (isPassenger && in.CallerId != conv.PassengerId) {
		return nil, status.Errorf(codes.PermissionDenied, "无权限操作该会话")
	}

	if isDriver {
		_ = l.svcCtx.DB.Model(&model.IMConversation{}).Where("id = ?", conv.Id).Update("unread_driver", 0)
	} else {
		_ = l.svcCtx.DB.Model(&model.IMConversation{}).Where("id = ?", conv.Id).Update("unread_passenger", 0)
	}
	return &__proto.MarkReadResponse{Ok: true}, nil
}
