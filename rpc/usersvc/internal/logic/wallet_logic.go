package logic

import (
	"context"
	"fmt"
	"time"

	"XiaoLong-Ridy/rpc/usersvc/internal/model"
	"XiaoLong-Ridy/rpc/usersvc/internal/svc"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
)

// WalletLogic 封装用户钱包余额查询、充值和提现流程。
type WalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWalletLogic 创建钱包业务逻辑实例。
func NewWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WalletLogic {
	return &WalletLogic{ctx: ctx, svcCtx: svcCtx}
}

// GetWallet 查询钱包余额及最近流水。
func (l *WalletLogic) GetWallet(req *userproto.GetWalletRequest) (*userproto.GetWalletResponse, error) {
	if req == nil || req.UserId == 0 || l.svcCtx == nil || l.svcCtx.Wallets == nil {
		return nil, fmt.Errorf("wallet repository not configured")
	}
	wallet, err := l.svcCtx.Wallets.Get(l.ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	list, total, err := l.svcCtx.Wallets.ListTransactions(l.ctx, req.UserId, 100, 0)
	if err != nil {
		return nil, err
	}
	return &userproto.GetWalletResponse{UserId: req.UserId, Balance: wallet.Balance, Transactions: toProtoTransactions(list), Total: total}, nil
}

// RechargeWallet 充值并在同一事务写入余额与充值流水。
func (l *WalletLogic) RechargeWallet(req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error) {
	return l.change(req, "recharge", "钱包充值", req.GetAmount())
}

// WithdrawWallet 提现并在同一事务写入余额与提现流水。
func (l *WalletLogic) WithdrawWallet(req *userproto.ChangeWalletRequest) (*userproto.ChangeWalletResponse, error) {
	return l.change(req, "withdraw", "钱包提现", -req.GetAmount())
}

func (l *WalletLogic) change(req *userproto.ChangeWalletRequest, typ, title string, amount float64) (*userproto.ChangeWalletResponse, error) {
	if req == nil || req.UserId == 0 || req.Amount <= 0 || req.Amount > 1000000 || l.svcCtx == nil || l.svcCtx.Wallets == nil {
		return nil, fmt.Errorf("invalid wallet change")
	}
	if err := l.svcCtx.Wallets.Change(l.ctx, req.UserId, typ, title, amount, 0); err != nil {
		return nil, err
	}
	wallet, err := l.svcCtx.Wallets.Get(l.ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	item := &userproto.WalletTransaction{UserId: req.UserId, Type: typ, Amount: amount, Title: title, CreatedAt: time.Now().Unix()}
	return &userproto.ChangeWalletResponse{Success: true, Balance: wallet.Balance, Transaction: item}, nil
}

func toProtoTransactions(list []*model.UserWalletTransaction) []*userproto.WalletTransaction {
	result := make([]*userproto.WalletTransaction, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		result = append(result, &userproto.WalletTransaction{Id: item.ID, UserId: item.UserID, Type: item.Type, Amount: item.Amount, OrderId: item.OrderID, Title: item.Title, CreatedAt: item.CreatedAt.Unix()})
	}
	return result
}
