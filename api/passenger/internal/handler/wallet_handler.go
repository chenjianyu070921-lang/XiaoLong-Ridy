package handler

import (
	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	"net/http"
)

// GetWalletHandler 查询当前乘客钱包。
func GetWalletHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := logic.NewWalletLogic(r.Context(), svcCtx, bearerToken(r)).GetWallet()
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// RechargeWalletHandler 处理充值请求。
func RechargeWalletHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WalletChangeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewWalletLogic(r.Context(), svcCtx, bearerToken(r)).RechargeWallet(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// WithdrawWalletHandler 处理提现请求。
func WithdrawWalletHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WalletChangeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewWalletLogic(r.Context(), svcCtx, bearerToken(r)).WithdrawWallet(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
