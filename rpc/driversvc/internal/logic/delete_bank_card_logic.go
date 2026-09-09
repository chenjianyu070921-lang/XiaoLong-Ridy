package logic

import (
	"context"
	
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	__proto "XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteBankCardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBankCardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBankCardLogic {
	return &DeleteBankCardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteBankCard 软删除司机银行卡：仅允许删除本人名下卡片。
func (l *DeleteBankCardLogic) DeleteBankCard(in *__proto.DeleteBankCardRequest) (*__proto.CommonResponse, error) {
	if in == nil || in.GetDriverId() <= 0 || in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverBankCardRepository == nil {
		return nil, status.Error(codes.InvalidArgument, "driver bank card repository not ready")
	}
	card, err := l.svcCtx.DriverBankCardRepository.GetByID(l.ctx, uint64(in.GetId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "银行卡不存在")
	}
	if card.DriverId != uint64(in.GetDriverId()) {
		return nil, status.Error(codes.InvalidArgument, "无权操作该银行卡")
	}
	if err := l.svcCtx.DriverBankCardRepository.Delete(l.ctx, uint64(in.GetId())); err != nil {
		return nil, err
	}
	return &__proto.CommonResponse{Message: "ok"}, nil
}
