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

// UpdateDriverLogic 处理更新司机请求的逻辑结构体。
type UpdateDriverLogic struct {
	ctx    context.Context      // ctx：请求上下文
	svcCtx *svc.ServiceContext  // svcCtx：服务上下文，持有 DB 等依赖
	logx.Logger
}

// NewUpdateDriverLogic 构造 UpdateDriverLogic 实例。
func NewUpdateDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDriverLogic {
	return &UpdateDriverLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateDriver 更新司机信息，仅修改请求中显式传入的字段。
func (l *UpdateDriverLogic) UpdateDriver(in *proto.UpdateDriverRequest) (*proto.UpdateDriverResponse, error) {
	// 先按 ID 查询司机是否存在
	var d model.Driver
	err := l.svcCtx.DB.First(&d, in.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("driver not found") // 司机不存在
	}
	if err != nil {
		return nil, err
	}

	// 仅更新请求中显式提供的字段（optional 字段为指针，nil 表示不更新）
	updates := map[string]interface{}{}
	if in.Phone != nil {
		updates["phone"] = in.GetPhone() // 手机号
	}
	if in.PasswordHash != nil {
		updates["password_hash"] = in.GetPasswordHash() // 密码哈希
	}
	if in.RealName != nil {
		updates["real_name"] = in.GetRealName() // 真实姓名
	}
	if in.IdCardNo != nil {
		updates["id_card_no"] = in.GetIdCardNo() // 身份证号
	}
	if in.DriverLicenseNo != nil {
		updates["driver_license_no"] = in.GetDriverLicenseNo() // 驾驶证号
	}
	if in.AvatarUrl != nil {
		updates["avatar_url"] = in.GetAvatarUrl() // 头像地址
	}
	if in.Status != nil {
		updates["status"] = int8(in.GetStatus()) // 账号状态
	}

	// 执行更新
	if err := l.svcCtx.DB.Model(&d).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新读取更新后的记录，获取最新状态与时间
	if err := l.svcCtx.DB.First(&d, in.Id).Error; err != nil {
		return nil, err
	}
	// 返回更新后的 ID、状态与更新时间
	return &proto.UpdateDriverResponse{
		Id:        int64(d.Id),
		Status:    proto.DriverStatus(d.Status),
		UpdatedAt: d.UpdatedAt.Unix(),
	}, nil
}
