package handler

import (
	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"
	"net/http"
)

// AvatarUploadTokenHandler 处理获取乘客头像七牛云上传凭证的请求。
func AvatarUploadTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AvatarUploadTokenRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		resp, err := logic.NewAvatarUploadLogic(r.Context(), svcCtx, bearerToken(r)).GetToken(&req)
		if err != nil {
			writeBusinessError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}
