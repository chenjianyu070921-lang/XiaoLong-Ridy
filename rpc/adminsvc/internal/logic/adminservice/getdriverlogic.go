package adminservicelogic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/repository"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDriverLogic {
	return &GetDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDriverLogic) GetDriver(in *adminsvc.DriverDetailRequest) (*adminsvc.Driver, error) {
	if in.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "driver id cannot be empty")
	}
	driver, err := l.svcCtx.DriverQueryRepository.Get(l.ctx, in.GetId())
	if errors.Is(err, repository.ErrDriverNotFound) {
		return nil, status.Error(codes.NotFound, "driver not found")
	}
	return driver, err
}
