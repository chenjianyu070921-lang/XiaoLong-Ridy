package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCouponLogic {
	return &UpdateCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新优惠券模板。
func (l *UpdateCouponLogic) UpdateCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
	// todo: add your logic here and delete this line

	return &adminsvc.CommonResponse{}, nil
}
