package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	var d model.Driver
	err := l.svcCtx.DB.First(&d, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("driver not found")
	}
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
			CreatedAt:       d.CreatedAt.Unix(),
			UpdatedAt:       d.UpdatedAt.Unix(),
		},
	}, nil
}
