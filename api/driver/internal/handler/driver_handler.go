package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// RegisterDriverHandler handles driver self-registration.
func RegisterDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterDriverRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).RegisterDriver(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// CreateDriverHandler handles backend-created driver accounts.
func CreateDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateDriverRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).CreateDriver(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// UpdateDriverHandler updates driver profile fields.
func UpdateDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateDriverRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if !middleware.IsInternalCall(r.Context()) {
			// 司机只能修改自己的资料：强制用 JWT 身份覆盖请求中的 id，防止越权修改其他司机（含改密码/手机号接管账号）。
			// 审核状态 status 只能由管理员/内部服务变更，司机端一律忽略，防止绕过审核或自行解冻。
			claims := middleware.ClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
				return
			}
			req.ID = int64(claims.AccountID)
			req.Status = nil
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).UpdateDriver(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

func UploadDriverAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.UploadDriverAvatarRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAvatarLogic().UploadDriverAvatar(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetDriverHandler returns the current driver profile. driverId comes from JWT.
func GetDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		driverID := int64(0)
		if middleware.IsInternalCall(r.Context()) {
			id, ok := decodeQueryID(r, "id")
			if !ok {
				writeError(w, http.StatusBadRequest, 50000, "invalid driverId")
				return
			}
			driverID = id
		} else {
			claims := middleware.ClaimsFromContext(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
				return
			}
			driverID = int64(claims.AccountID)
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).GetDriver(driverID)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetDriverByPhoneHandler returns a driver profile by phone number.
func GetDriverByPhoneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.GetDriverByPhoneRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).GetDriverByPhone(req.Phone)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ListNearbyDriversHandler returns nearby online drivers for dispatching.
func ListNearbyDriversHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !middleware.IsInternalCall(r.Context()) && middleware.ClaimsFromContext(r.Context()) == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.ListNearbyDriversRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).ListNearbyDrivers(&req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetOrderHeatmapHandler returns nearby wait-accept order heat points for the current driver.
func GetOrderHeatmapHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		var req types.HeatmapRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewHeatmapLogic(r.Context(), svcCtx).GetOrderHeatmap(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// DeleteDriverHandler deletes a driver by id. This route is not used by driver web.
func DeleteDriverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := decodeQueryID(r, "id")
		if !ok {
			writeError(w, http.StatusBadRequest, 50000, "司机ID不合法")
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).DeleteDriver(id)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetDriverAiScoreHandler returns the current driver's AI score.
func GetDriverAiScoreHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "login credential invalid")
			return
		}
		resp, err := logic.NewDriverLogic(r.Context(), svcCtx).GetDriverAiScore(int64(claims.AccountID))
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
