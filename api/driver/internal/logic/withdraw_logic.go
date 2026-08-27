package logic

import (
	"context"

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

func (l *WithdrawLogic) CreateWithdraw(driverID int64, req *types.CreateWithdrawRequest) (*types.CreateWithdrawResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.CreateWithdraw(l.ctx, &driversproto.CreateWithdrawRequest{
		DriverId:   driverID,
		Amount:     req.Amount,
		PayeeName:  req.PayeeName,
		PayAccount: req.PayAccount,
	})
	if err != nil {
		return nil, err
	}
	return &types.CreateWithdrawResponse{
		ID:         resp.GetId(),
		WithdrawNo: resp.GetWithdrawNo(),
		Status:     resp.GetStatus(),
		CreatedAt:  resp.GetCreatedAt(),
	}, nil
}

func (l *WithdrawLogic) ListWithdraws(driverID int64, req *types.ListWithdrawsRequest) (*types.ListWithdrawsResponse, error) {
	if driverID <= 0 || req == nil {
		return nil, ErrInvalidParam
	}
	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}
	page, pageSize := clampPage(req.Page, req.PageSize)
	resp, err := client.ListWithdraws(l.ctx, &driversproto.ListWithdrawsRequest{
		DriverId: driverID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	result := &types.ListWithdrawsResponse{Total: resp.GetTotal()}
	for _, record := range resp.GetRecords() {
		result.Records = append(result.Records, types.WithdrawRecord{
			ID:         record.GetId(),
			WithdrawNo: record.GetWithdrawNo(),
			Amount:     record.GetAmount(),
			PayeeName:  record.GetPayeeName(),
			PayAccount: record.GetPayAccount(),
			Status:     record.GetStatus(),
			Remark:     record.GetRemark(),
			AppliedAt:  record.GetAppliedAt(),
			PaidAt:     record.GetPaidAt(),
			CreatedAt:  record.GetCreatedAt(),
		})
	}
	return result, nil
}

func (l *WithdrawLogic) driverClient() (svc.DriverClient, error) {
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	return l.svcCtx.DriverClient, nil
}
