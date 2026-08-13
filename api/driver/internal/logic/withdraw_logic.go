// Package logic 实现 driver API 的业务逻辑层。
package logic

import (
	"context" // 用于在不同层之间传递请求上下文
	"errors"   // 用于返回业务校验错误

	"XiaoLong-Ridy/api/driver/internal/svc"    // 服务上下文，提供 driversvc 客户端
	"XiaoLong-Ridy/api/driver/internal/types"  // API 层使用的请求/响应类型
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto" // driversvc 的 gRPC 请求/响应类型
)

// WithdrawLogic 提现业务逻辑处理器，持有上下文与下游客户端。
type WithdrawLogic struct {
	ctx    context.Context    // 当前请求上下文
	svcCtx *svc.ServiceContext // 全局服务上下文（含 driversvc 客户端）
}

// NewWithdrawLogic 构造提现逻辑处理器实例。
func NewWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawLogic {
	// 注入上下文与服务上下文。
	return &WithdrawLogic{ctx: ctx, svcCtx: svcCtx}
}

// CreateWithdraw 创建提现，校验司机归属、单号、金额与收款信息。
func (l *WithdrawLogic) CreateWithdraw(req *types.CreateWithdrawRequest) (*types.CreateWithdrawResponse, error) {
	// 校验所属司机 ID 合法性。
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	// 校验提现单号非空。
	if req.WithdrawNo == "" {
		return nil, errors.New("提现单号不能为空")
	}
	// 校验提现金额必须大于 0。
	if req.Amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}
	// 校验收款人姓名非空。
	if req.PayeeName == "" {
		return nil, errors.New("收款人姓名不能为空")
	}
	// 校验收款账号非空。
	if req.PayAccount == "" {
		return nil, errors.New("收款账号不能为空")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游创建提现接口。
	resp, err := client.CreateWithdraw(l.ctx, &driversproto.CreateWithdrawRequest{
		DriverId:   req.DriverID,   // 司机 ID
		WithdrawNo: req.WithdrawNo, // 提现单号
		Amount:     req.Amount,     // 提现金额
		PayeeName:  req.PayeeName,  // 收款人姓名
		PayAccount: req.PayAccount, // 收款账号
	})
	if err != nil {
		return nil, err
	}
	// 返回创建结果（提现 ID + 单号 + 初始状态）。
	return &types.CreateWithdrawResponse{ID: resp.GetId(), WithdrawNo: resp.GetWithdrawNo(), Status: resp.GetStatus()}, nil
}

// UpdateWithdraw 更新提现，校验 ID 与可选金额/状态范围。
func (l *WithdrawLogic) UpdateWithdraw(req *types.UpdateWithdrawRequest) (*types.UpdateWithdrawResponse, error) {
	// 校验提现记录 ID 合法性。
	if req.ID <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	// 若传入金额，校验其大于 0。
	if req.Amount != nil && *req.Amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}
	// 若传入状态，校验其范围 1~3。
	if req.Status != nil && (*req.Status < 1 || *req.Status > 3) {
		return nil, errors.New("提现状态不合法(1申请中 2打款成功 3打款失败)")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游更新接口，可选字段直接透传指针。
	resp, err := client.UpdateWithdraw(l.ctx, &driversproto.UpdateWithdrawRequest{
		Id:          req.ID,          // 提现记录 ID
		DriverId:    req.DriverID,    // 可选司机 ID
		WithdrawNo:  req.WithdrawNo,  // 可选提现单号
		Amount:      req.Amount,      // 可选金额
		PayeeName:   req.PayeeName,   // 可选收款人
		PayAccount:  req.PayAccount,  // 可选收款账号
		Status:      req.Status,      // 可选状态
		Remark:      req.Remark,      // 可选备注
		AppliedAt:   req.AppliedAt,   // 可选申请时间
		PaidAt:      req.PaidAt,      // 可选打款时间
	})
	if err != nil {
		return nil, err
	}
	// 返回更新结果。
	return &types.UpdateWithdrawResponse{ID: resp.GetId(), Status: resp.GetStatus()}, nil
}

// GetWithdraw 查询提现详情。
func (l *WithdrawLogic) GetWithdraw(id int64) (*types.GetWithdrawResponse, error) {
	// 校验提现记录 ID 合法性。
	if id <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游查询接口。
	resp, err := client.GetWithdraw(l.ctx, &driversproto.GetWithdrawRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 取出 proto 中的提现实体。
	w := resp.GetWithdraw()
	// 映射为 API 的提现详情结构并返回。
	return &types.GetWithdrawResponse{Withdraw: types.WithdrawDetail{
		ID:          w.GetId(),
		DriverID:    w.GetDriverId(),
		WithdrawNo:  w.GetWithdrawNo(),
		Amount:      w.GetAmount(),
		PayeeName:   w.GetPayeeName(),
		PayAccount:  w.GetPayAccount(),
		Status:      w.GetStatus(),
		Remark:      w.GetRemark(),
		AppliedAt:   w.GetAppliedAt(),
		PaidAt:      w.GetPaidAt(),
		CreatedAt:   w.GetCreatedAt(),
	}}, nil
}

// DeleteWithdraw 删除提现记录。
func (l *WithdrawLogic) DeleteWithdraw(id int64) (*types.DeleteResponse, error) {
	// 校验提现记录 ID 合法性。
	if id <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游删除接口。
	resp, err := client.DeleteWithdraw(l.ctx, &driversproto.DeleteWithdrawRequest{Id: id})
	if err != nil {
		return nil, err
	}
	// 返回删除结果。
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

// ListWithdraws 分页查询提现列表。
func (l *WithdrawLogic) ListWithdraws(req *types.ListWithdrawsRequest) (*types.ListWithdrawsResponse, error) {
	// 收敛分页参数到合法范围。
	page, pageSize := clampPage(req.Page, req.PageSize)
	// 获取 driversvc 客户端。
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	// 调用下游列表接口。
	resp, err := client.ListWithdraws(l.ctx, &driversproto.ListWithdrawsRequest{
		DriverId: req.DriverID, Status: req.Status, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	// 预分配切片。
	list := make([]types.WithdrawSummary, 0, len(resp.GetList()))
	// 遍历并映射为 API 摘要结构。
	for _, s := range resp.GetList() {
		list = append(list, types.WithdrawSummary{
			ID:         s.GetId(),
			DriverID:   s.GetDriverId(),
			WithdrawNo: s.GetWithdrawNo(),
			Amount:     s.GetAmount(),
			Status:     s.GetStatus(),
			AppliedAt:  s.GetAppliedAt(),
		})
	}
	// 组装分页响应返回。
	return &types.ListWithdrawsResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

// client 从服务上下文中安全取出 driversvc 客户端。
func (l *WithdrawLogic) client() (svc.DriverClient, error) {
	// 防御性校验客户端是否可用。
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
