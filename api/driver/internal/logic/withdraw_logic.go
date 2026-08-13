package logic

import (
	"context"
	"errors"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
)

type WithdrawLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawLogic {
	return &WithdrawLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *WithdrawLogic) CreateWithdraw(req *types.CreateWithdrawRequest) (*types.CreateWithdrawResponse, error) {
	if req.DriverID <= 0 {
		return nil, errors.New("司机ID不合法")
	}
	if req.WithdrawNo == "" {
		return nil, errors.New("提现单号不能为空")
	}
	if req.Amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}
	if req.PayeeName == "" {
		return nil, errors.New("收款人姓名不能为空")
	}
	if req.PayAccount == "" {
		return nil, errors.New("收款账号不能为空")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateWithdraw(l.ctx, &driversproto.CreateWithdrawRequest{
		DriverId:   req.DriverID,
		WithdrawNo: req.WithdrawNo,
		Amount:     req.Amount,
		PayeeName:  req.PayeeName,
		PayAccount: req.PayAccount,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateWithdrawResponse{ID: resp.GetId(), WithdrawNo: resp.GetWithdrawNo(), Status: resp.GetStatus()}, nil
}

func (l *WithdrawLogic) UpdateWithdraw(req *types.UpdateWithdrawRequest) (*types.UpdateWithdrawResponse, error) {
	if req.ID <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	if req.Amount != nil && *req.Amount <= 0 {
		return nil, errors.New("提现金额必须大于0")
	}
	if req.Status != nil && (*req.Status < 1 || *req.Status > 3) {
		return nil, errors.New("提现状态不合法(1申请中 2打款成功 3打款失败)")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.UpdateWithdraw(l.ctx, &driversproto.UpdateWithdrawRequest{
		Id:          req.ID,
		DriverId:    req.DriverID,
		WithdrawNo:  req.WithdrawNo,
		Amount:      req.Amount,
		PayeeName:   req.PayeeName,
		PayAccount:  req.PayAccount,
		Status:      req.Status,
		Remark:      req.Remark,
		AppliedAt:   req.AppliedAt,
		PaidAt:      req.PaidAt,
	})
	if err != nil {
		return nil, err
	}
	return &types.UpdateWithdrawResponse{ID: resp.GetId(), Status: resp.GetStatus()}, nil
}

func (l *WithdrawLogic) GetWithdraw(id int64) (*types.GetWithdrawResponse, error) {
	if id <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetWithdraw(l.ctx, &driversproto.GetWithdrawRequest{Id: id})
	if err != nil {
		return nil, err
	}
	w := resp.GetWithdraw()
	return &types.GetWithdrawResponse{Withdraw: types.WithdrawDetail{
		ID: w.GetId(), DriverID: w.GetDriverId(), WithdrawNo: w.GetWithdrawNo(), Amount: w.GetAmount(),
		PayeeName: w.GetPayeeName(), PayAccount: w.GetPayAccount(), Status: w.GetStatus(), Remark: w.GetRemark(),
		AppliedAt: w.GetAppliedAt(), PaidAt: w.GetPaidAt(), CreatedAt: w.GetCreatedAt(),
	}}, nil
}

func (l *WithdrawLogic) DeleteWithdraw(id int64) (*types.DeleteResponse, error) {
	if id <= 0 {
		return nil, errors.New("提现记录ID不合法")
	}
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.DeleteWithdraw(l.ctx, &driversproto.DeleteWithdrawRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return &types.DeleteResponse{ID: resp.GetId(), Success: resp.GetSuccess()}, nil
}

func (l *WithdrawLogic) ListWithdraws(req *types.ListWithdrawsRequest) (*types.ListWithdrawsResponse, error) {
	page, pageSize := clampPage(req.Page, req.PageSize)
	client, err := l.client()
	if err != nil {
		return nil, err
	}
	resp, err := client.ListWithdraws(l.ctx, &driversproto.ListWithdrawsRequest{
		DriverId: req.DriverID, Status: req.Status, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]types.WithdrawSummary, 0, len(resp.GetList()))
	for _, s := range resp.GetList() {
		list = append(list, types.WithdrawSummary{
			ID: s.GetId(), DriverID: s.GetDriverId(), WithdrawNo: s.GetWithdrawNo(),
			Amount: s.GetAmount(), Status: s.GetStatus(), AppliedAt: s.GetAppliedAt(),
		})
	}
	return &types.ListWithdrawsResponse{List: list, Total: resp.GetTotal(), Page: resp.GetPage(), PageSize: resp.GetPageSize()}, nil
}

func (l *WithdrawLogic) client() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
