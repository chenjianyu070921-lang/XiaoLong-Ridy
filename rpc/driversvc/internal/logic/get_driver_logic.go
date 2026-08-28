package logic

import (
	"context"
	"errors"

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
	if in == nil || in.Id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	d, err := l.svcCtx.DriverRepository.GetByID(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}
	return &proto.GetDriverResponse{
		Driver: &proto.Driver{
			Id:              int64(d.Id),
			Phone:           d.Phone,
			PasswordHash:    d.PasswordHash,
			RealName:        d.RealName,
			IdCardNo:        d.IdCardNo,
			DriverLicenseNo: d.DriverLicenseNo,
			AvatarUrl:       d.AvatarUrl,
			Status:          proto.DriverStatus(d.Status),
			OnlineStatus:    int32(d.OnlineStatus),
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}
