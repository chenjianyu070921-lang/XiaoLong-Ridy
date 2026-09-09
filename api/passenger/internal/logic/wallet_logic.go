package logic

import (
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	userproto "XiaoLong-Ridy/rpc/usersvc/proto"
	"context"
)

// WalletLogic 负责乘客钱包接口的登录态校验和 usersvc RPC 编排。
type WalletLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	token  string
}

// NewWalletLogic 创建钱包逻辑实例。
func NewWalletLogic(ctx context.Context, svcCtx *svc.ServiceContext, token string) *WalletLogic {
	return &WalletLogic{ctx: ctx, svcCtx: svcCtx, token: token}
}

// GetWallet 查询后端 MySQL 钱包余额和流水。
func (l *WalletLogic) GetWallet() (*types.WalletResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.GetWallet(l.ctx, &userproto.GetWalletRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	result := &types.WalletResponse{UserID: resp.GetUserId(), Balance: resp.GetBalance(), Total: resp.GetTotal(), Transactions: make([]types.WalletTransaction, 0, len(resp.GetTransactions()))}
	for _, item := range resp.GetTransactions() {
		if item == nil {
			continue
		}
		result.Transactions = append(result.Transactions, types.WalletTransaction{ID: item.GetId(), Type: item.GetType(), Amount: item.GetAmount(), OrderID: item.GetOrderId(), Title: item.GetTitle(), CreatedAt: item.GetCreatedAt()})
	}
	return result, nil
}

// RechargeWallet 请求后端执行充值并写入 MySQL 钱包流水。
func (l *WalletLogic) RechargeWallet(req *types.WalletChangeRequest) (*types.WalletChangeResponse, error) {
	return l.change(req, true)
}

// WithdrawWallet 请求后端执行提现并写入 MySQL 钱包流水。
func (l *WalletLogic) WithdrawWallet(req *types.WalletChangeRequest) (*types.WalletChangeResponse, error) {
	return l.change(req, false)
}

func (l *WalletLogic) change(req *types.WalletChangeRequest, recharge bool) (*types.WalletChangeResponse, error) {
	userID, err := currentUserID(l.svcCtx, l.token)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Amount <= 0 {
		return nil, ErrInvalidRequest
	}
	client, err := l.userClient()
	if err != nil {
		return nil, err
	}
	rpcReq := &userproto.ChangeWalletRequest{UserId: userID, Amount: req.Amount}
	var resp *userproto.ChangeWalletResponse
	if recharge {
		resp, err = client.RechargeWallet(l.ctx, rpcReq)
	} else {
		resp, err = client.WithdrawWallet(l.ctx, rpcReq)
	}
	if err != nil {
		return nil, err
	}
	item := types.WalletTransaction{}
	if resp.GetTransaction() != nil {
		item = types.WalletTransaction{ID: resp.GetTransaction().GetId(), Type: resp.GetTransaction().GetType(), Amount: resp.GetTransaction().GetAmount(), OrderID: resp.GetTransaction().GetOrderId(), Title: resp.GetTransaction().GetTitle(), CreatedAt: resp.GetTransaction().GetCreatedAt()}
	}
	return &types.WalletChangeResponse{Success: resp.GetSuccess(), Balance: resp.GetBalance(), Transaction: item}, nil
}

func (l *WalletLogic) userClient() (svc.UserClient, error) {
	if l.svcCtx == nil || l.svcCtx.UserClient == nil {
		return nil, ErrUserClientNotConfigured
	}
	return l.svcCtx.UserClient, nil
}
