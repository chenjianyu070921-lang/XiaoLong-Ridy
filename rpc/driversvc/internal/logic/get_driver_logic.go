package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
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

// GetDriver 根据司机 ID 查询司机完整信息。
func (l *GetDriverLogic) GetDriver(in *proto.GetDriverRequest) (*proto.GetDriverResponse, error) {
	d, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	driver, err := buildAdminDriverPB(l.ctx, l.svcCtx, d)
	if err != nil {
		return nil, err
	}
	return &proto.GetDriverResponse{
		Driver: driver,
	}, nil
}
