// Package logic 实现 driver API 的业务逻辑层。
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
	ctx    context.Context   // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewDriverLogic 构造司机逻辑处理器实例。
func NewDriverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DriverLogic {
	// 注入上下文与服务上下文。
	return &DriverLogic{ctx: ctx, svcCtx: svcCtx}
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

// ListDrivers 分页查询司机列表。
func (l *DriverLogic) ListDrivers(req *types.ListDriversRequest) (*types.ListDriversResponse, error) {
	// 收敛分页参数到合法范围。
	page, pageSize := clampPage(req.Page, req.PageSize)
	// 获取 driversvc 客户端。
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	// 调用下游列表接口，状态过滤经枚举转换。
	resp, err := client.ListDrivers(l.ctx, &driversproto.ListDriversRequest{
		Status:       enumDriverStatusStr(req.Status), // 状态字符串转枚举
		PhoneKeyword: req.PhoneKeyword,                // 手机号模糊关键字
		Page:         page,                            // 页码
		PageSize:     pageSize,                        // 每页条数
	})
	if err != nil {
		return nil, err
	}
	// 预分配切片，避免扩容。
	list := make([]types.DriverSummary, 0, len(resp.GetList()))
	// 遍历 proto 摘要列表，逐个映射为 API 摘要结构。
	for _, s := range resp.GetList() {
		list = append(list, types.DriverSummary{
			ID:              s.GetId(),
			Phone:           s.GetPhone(),
			RealName:        s.GetRealName(),
			DriverLicenseNo: s.GetDriverLicenseNo(),
			Status:          s.GetStatus().String(),
			CreatedAt:       s.GetCreatedAt(),
		})
	}
	// 组装并返回分页响应。
	return &types.ListDriversResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
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
