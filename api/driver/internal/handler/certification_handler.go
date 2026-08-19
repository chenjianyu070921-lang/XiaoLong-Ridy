package handler

import (
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/middleware"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

// UploadCertificationHandler POST /api/driver/v1/drivers/certification/upload
// 司机资质上传：从 JWT 取司机 ID，解析 base64 图片，调用逻辑直传 MinIO 并落库，返回资质记录与审核状态。
// 需携带有效 JWT。
func UploadCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		var req types.UploadCertificationRequest
		decodeJSON(w, r, &req)
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).UploadCertification(int64(claims.AccountID), &req)
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// GetCertificationHandler GET /api/driver/v1/drivers/certification
// 司机资质查询：从 JWT 取司机 ID，返回其资质记录与审核状态。需携带有效 JWT。
func GetCertificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "登录凭证无效")
			return
		}
		resp, err := logic.NewCertificationLogic(r.Context(), svcCtx).GetCertification(int64(claims.AccountID))
		if err != nil {
			writeParamError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
