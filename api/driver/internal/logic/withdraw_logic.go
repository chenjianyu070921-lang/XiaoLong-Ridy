package logic

import (
	"context"
	"errors"
	"math"
	"strings"

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
	if l.svcCtx == nil || l.svcCtx.DriverClient == nil {
		return nil, ErrDriverClientNotConfigured
	}
	if err := validateWithdrawAmount(req.Amount); err != nil {
		return nil, err
	}

	payeeName := strings.TrimSpace(req.PayeeName)
	if payeeName == "" {
		return nil, errors.New("payee name is required")
	}
	payAccount := strings.TrimSpace(req.PayAccount)
	if payAccount == "" {
		return nil, errors.New("pay account is required")
	}

	client, err := l.driverClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.CreateWithdraw(l.ctx, &driversproto.CreateWithdrawRequest{
		DriverId:   driverID,
		Amount:     normalizeWithdrawAmount(req.Amount),
		PayeeName:  payeeName,
		PayAccount: payAccount,
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
		result.List = append(result.List, types.WithdrawRecord{
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

func validateWithdrawAmount(amount float64) error {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return errors.New("withdraw amount must be a finite number")
	}
	if amount <= 0 {
		return errors.New("withdraw amount must be greater than 0")
	}
	if amount > 100000 {
		return errors.New("withdraw amount exceeds limit")
	}
	cents := amount * 100
	if math.Abs(cents-math.Round(cents)) > 1e-6 {
		return errors.New("withdraw amount must not exceed 2 decimal places")
	}
	return nil
}

func normalizeWithdrawAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}
