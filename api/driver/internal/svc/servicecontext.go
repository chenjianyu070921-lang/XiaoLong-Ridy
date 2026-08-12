// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"driver/internal/config"
	"usersvc/usersvcclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config  config.Config
	UserRpc usersvcclient.Usersvc
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		UserRpc: usersvcclient.NewUsersvc(zrpc.MustNewClient(c.UserRpc)),
	}
}
