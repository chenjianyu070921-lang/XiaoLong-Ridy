package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// SendBankCardSmsCodeHandler POST /api/driver/v1/bank-cards/sms-code
func SendBankCardSmsCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendBankCardSmsCodeRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		err := logic.NewBankCardLogic(r.Context(), svcCtx).SendBankCardSmsCode(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, map[string]bool{"success": true})
	}
}

// BindBankCardHandler POST /api/driver/v1/bank-cards
func BindBankCardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.BindBankCardRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewBankCardLogic(r.Context(), svcCtx).BindBankCard(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListBankCardsHandler GET /api/driver/v1/bank-cards
func ListBankCardsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		resp, err := logic.NewBankCardLogic(r.Context(), svcCtx).ListBankCards(int64(claims.AccountID))
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// DeleteBankCardHandler POST /api/driver/v1/bank-cards/delete
func DeleteBankCardHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.DeleteBankCardRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		err := logic.NewBankCardLogic(r.Context(), svcCtx).DeleteBankCard(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, map[string]bool{"success": true})
	}
}

// ResetWithdrawPasswordHandler POST /api/driver/v1/bank-cards/reset-password
func ResetWithdrawPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.ResetWithdrawPasswordRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewBankCardLogic(r.Context(), svcCtx).ResetWithdrawPassword(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
