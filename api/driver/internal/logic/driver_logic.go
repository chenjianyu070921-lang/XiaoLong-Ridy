// Package logic 实现 driver API 的司机业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"   // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"  // API 层使用的请求/响应类型
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// DriverLogic 司机业务逻辑处理器，持有上下文与下游客户端。
type DriverLogic struct {
	ctx    context.Context    // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewDriverLogic 构造司机逻辑处理器实例。
func NewDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLogic {
	// 注入上下文与服务上下文。
	return &DriverLogic{ctx: ctx, svcCtx: svcCtx}
}

// enumDriverStatus 将可选状态字符串转为 proto 枚举指针；nil 或空串返回 nil（表示不更新该字段）。
func enumDriverStatus(s *string) *driversproto.DriverStatus {
	// 入参为 nil 或空串时返回 nil，调用方据此跳过该可选字段。
	if s == nil || *s == "" {
		return nil
	}
	// 声明局部变量承载映射后的枚举值。
	var v driversproto.DriverStatus
	// 按字符串值映射到对应的 proto 枚举。
	switch *s {
	case "DRIVER_STATUS_PENDING": // 待审核
		v = driversproto.DriverStatus_DRIVER_STATUS_PENDING
	case "DRIVER_STATUS_NORMAL": // 正常
		v = driversproto.DriverStatus_DRIVER_STATUS_NORMAL
	case "DRIVER_STATUS_FROZEN": // 冻结
		v = driversproto.DriverStatus_DRIVER_STATUS_FROZEN
	case "DRIVER_STATUS_CANCELLED": // 注销
		v = driversproto.DriverStatus_DRIVER_STATUS_CANCELLED
	default: // 未知值映射为未指定（由底层忽略）。
		v = driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED
	}
	// 返回枚举值的指针，以匹配 proto 的 optional 字段语义。
	return &v
}

// CreateDriver 创建司机，校验必填项与手机号/身份证格式。
func (l *DriverLogic) CreateDriver(req *types.CreateDriverRequest) (*types.CreateDriverResponse, error) {
	// 校验手机号格式，不合法直接返回错误。
	if !validPhone(req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	// 校验真实姓名非空。
	if req.RealName == "" {
		return nil, errors.New("真实姓名不能为空")
	}
	// 校验身份证号格式。
	if !validIDCard(req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	// 校验驾驶证号非空。
	if req.DriverLicenseNo == "" {
		return nil, errors.New("驾驶证号不能为空")
	}
	// 获取 driversvc 客户端（可能为配置错误）。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游创建司机接口，并将 API 入参映射为 proto 请求。
	resp, err := client.CreateDriver(l.ctx, &driversproto.CreateDriverRequest{
		Phone:           req.Phone,            // 手机号
		PasswordHash:    req.PasswordHash,     // 密码哈希
		RealName:        req.RealName,         // 真实姓名
		IdCardNo:        req.IdCardNo,         // 身份证号
		DriverLicenseNo: req.DriverLicenseNo,  // 驾驶证号
		AvatarUrl:       req.AvatarURL,        // 头像地址
	})
	if err != nil {
		// 下游调用失败，向上透传错误。
		return nil, err
	}
	// 将 proto 响应转换为 API 响应并返回。
	return &types.CreateDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), CreatedAt: resp.GetCreatedAt()}, nil
}

// UpdateDriver 更新司机，校验存在性后转发可选字段。
func (l *DriverLogic) UpdateDriver(req *types.UpdateDriverRequest) (*types.UpdateDriverResponse, error) {
	// 校验司机 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 若传入手机号，则校验其格式。
	if req.Phone != nil && !validPhone(*req.Phone) {
		return nil, errors.New("手机号格式不合法")
	}
	// 若传入身份证号，则校验其格式。
	if req.IdCardNo != nil && !validIDCard(*req.IdCardNo) {
		return nil, errors.New("身份证号格式不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口；可选字段直接透传指针，状态经枚举转换。
	resp, err := client.UpdateDriver(l.ctx, &driversproto.UpdateDriverRequest{
		Id:              req.ID,                       // 司机 ID
		Phone:           req.Phone,                   // 可选手机号
		PasswordHash:    req.PasswordHash,            // 可选密码哈希
		RealName:        req.RealName,                // 可选姓名
		IdCardNo:        req.IdCardNo,                // 可选身份证号
		DriverLicenseNo: req.DriverLicenseNo,         // 可选驾驶证号
		AvatarUrl:       req.AvatarURL,               // 可选头像
		Status:          enumDriverStatus(req.Status), // 可选状态（字符串转枚举指针）
	})
	if err != nil {
		return nil, err
	}
	// 转换响应并返回。
	return &types.UpdateDriverResponse{ID: resp.GetId(), Status: resp.GetStatus().String(), UpdatedAt: resp.GetUpdatedAt()}, nil
}

// GetDriver 查询司机详情。
func (l *DriverLogic) GetDriver(id int64) (*types.GetDriverResponse, error) {
	// 校验司机 ID 合法性。
	if id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetDriver(l.ctx, &driversproto.GetDriverRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的司机实体。
	d := resp.GetDriver()
	// 将 proto 实体映射为 API 的司机详情结构并返回。
	return &types.GetDriverResponse{Driver: types.DriverDetail{
		ID:              d.GetId(),
		Phone:           d.GetPhone(),
		RealName:        d.GetRealName(),
		IdCardNo:        d.GetIdCardNo(),
		DriverLicenseNo: d.GetDriverLicenseNo(),
		AvatarURL:       d.GetAvatarUrl(),
		Status:          d.GetStatus().String(),
		CreatedAt:       d.GetCreatedAt(),
		UpdatedAt:       d.GetUpdatedAt(),
	}}, nil
}

// DeleteDriver 删除（软删）司机。
func (l *DriverLogic) DeleteDriver(id int64) (*types.DeleteResponse, error) {
	// 校验司机 ID 合法性。
	if id <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游删除接口。
	resp, err := client.DeleteDriver(l.ctx, &driversproto.DeleteDriverRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 返回删除结果（ID + 是否成功）。
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// driverClient 从服务上下文中安全取出 driversvc 客户端。
func (l *DriverLogic) driverClient() (svc.DriverClient, error) {
	// 防御性校验：上下文或服务客户端为空时返回配置错误。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	// 返回已配置的客户端。
	return l.svcCtx.DriverClient, nil
}
