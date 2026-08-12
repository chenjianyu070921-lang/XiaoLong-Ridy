package logic

import (
	"context"

	"usersvc/internal/svc"
	"usersvc/usersvc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *usersvc.LoginReq) (*usersvc.LoginResp, error) {
	driver, err := l.svcCtx.DriverModel.FindOneByPhone(l.ctx, in.Phone)
	if err == sqlx.ErrNotFound {
		return &usersvc.LoginResp{Code: 2001, Message: "司机不存在"}, nil
	}
	if err != nil {
		logx.Errorf("find driver by phone failed: %v", err)
		return &usersvc.LoginResp{Code: 5000, Message: "系统错误"}, nil
	}

	return &usersvc.LoginResp{
		Code:     0,
		Message:  "success",
		DriverId: driver.Id,
		Status:   int32(driver.Status),
	}, nil
}
