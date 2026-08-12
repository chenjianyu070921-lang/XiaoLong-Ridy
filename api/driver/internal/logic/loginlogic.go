// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"fmt"
	"time"

	"driver/internal/svc"
	"driver/internal/types"
	"usersvc/usersvcclient"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 调用下游 usersvc 完成手机号+验证码登录校验
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &usersvcclient.LoginReq{
		Phone: req.Phone,
		Code:  req.Code,
	})
	if err != nil {
		logx.Errorf("user rpc login failed: %v", err)
		return nil, err
	}
	if rpcResp.Code != 0 {
		return nil, fmt.Errorf("login failed code=%d msg=%s", rpcResp.Code, rpcResp.Message)
	}

	// 签发 JWT
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"driverId": rpcResp.DriverId,
		"exp":      now + l.svcCtx.Config.Auth.AccessExpire,
		"iat":      now,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(l.svcCtx.Config.Auth.AccessSecret))
	if err != nil {
		logx.Errorf("sign jwt failed: %v", err)
		return nil, err
	}

	return &types.LoginResp{
		Token:    signed,
		DriverId: rpcResp.DriverId,
	}, nil
}
