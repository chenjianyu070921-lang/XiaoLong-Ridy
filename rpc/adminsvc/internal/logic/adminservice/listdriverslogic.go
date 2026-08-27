package adminservicelogic

import (
	"context"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/repository"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDriversLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListDriversLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDriversLogic {
	return &ListDriversLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListDriversLogic) ListDrivers(in *adminsvc.DriverListRequest) (*adminsvc.DriverListResponse, error) {
	result, err := l.svcCtx.DriverQueryRepository.List(l.ctx, repository.DriverListFilter{
		Page:     in.GetPage(),
		PageSize: in.GetPageSize(),
		Keyword:  in.GetKeyword(),
		Status:   in.GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	return &adminsvc.DriverListResponse{
		List:     result.List,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}
