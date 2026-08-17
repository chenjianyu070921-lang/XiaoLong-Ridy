package logic

import (
	"context"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDriverLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDriverLogic {
	return &CreateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateDriver 创建司机账号，状态初始为待审核（PENDING）。
func (l *CreateDriverLogic) CreateDriver(in *proto.CreateDriverRequest) (*proto.CreateDriverResponse, error) {
	d := &model.Driver{
		Phone:           in.Phone,
		PasswordHash:    in.PasswordHash,
		RealName:        in.RealName,
		IdCardNo:        in.IdCardNo,
		DriverLicenseNo: in.DriverLicenseNo,
		AvatarUrl:       in.AvatarUrl,
		Status:          int8(proto.DriverStatus_DRIVER_STATUS_PENDING),
	}
	if err := l.svcCtx.DriverRepository.Create(l.ctx, d); err != nil {
		return nil, err
	}
	return &proto.CreateDriverResponse{
		Id:        int64(d.Id),
		Status:    proto.DriverStatus_DRIVER_STATUS_PENDING,
		CreatedAt: d.CreatedAt.Unix(),
	}, nil
}
