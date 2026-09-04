package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/rpc/driversvc/internal/model"
	"XiaoLong-Ridy/rpc/driversvc/internal/svc"
	"XiaoLong-Ridy/rpc/driversvc/proto"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

// ErrDriverAlreadyExists 表示手机号或驾驶证号已存在（唯一索引冲突）。
var ErrDriverAlreadyExists = errors.New("driver already exists")

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
	if in == nil {
		return nil, errors.New("请求参数不能为空")
	}
	if l.svcCtx == nil || l.svcCtx.DriverRepository == nil {
		return nil, errors.New("driver repository not ready")
	}
	if err := validateDriverIdentity(in.Phone, in.RealName, in.IdCardNo, in.DriverLicenseNo); err != nil {
		return nil, err
	}
	passwordHash, err := prepareDriverPasswordHash(in.PasswordHash)
	if err != nil {
		return nil, err
	}
	d := &model.Driver{
		Phone:           in.Phone,
		PasswordHash:    passwordHash,
		RealName:        in.RealName,
		IdCardNo:        in.IdCardNo,
		DriverLicenseNo: in.DriverLicenseNo,
		AvatarUrl:       in.AvatarUrl,
		Status:          int8(proto.DriverStatus_DRIVER_STATUS_PENDING),
	}
	if err := l.svcCtx.DriverRepository.Create(l.ctx, d); err != nil {
		// 手机号/驾驶证号唯一索引冲突时，返回友好的业务错误而非裸 MySQL 错误。
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil, ErrDriverAlreadyExists
		}
		return nil, err
	}
	return &proto.CreateDriverResponse{
		Id:        int64(d.Id),
		Status:    proto.DriverStatus_DRIVER_STATUS_PENDING,
		CreatedAt: d.CreatedAt.Unix(),
	}, nil
}
