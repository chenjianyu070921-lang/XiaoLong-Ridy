package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableCouponLogic {
	return &DisableCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 下架优惠券模板。
func (l *DisableCouponLogic) DisableCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.CommonResponse{}, nil
}
