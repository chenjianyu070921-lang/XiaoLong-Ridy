package handler

import (
	"errors"
	"net/http"

	"XiaoLong-Ridy/api/driver/internal/logic"
	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/api/driver/internal/types"
)

func ImgCaptchaHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		phone := r.URL.Query().Get("phone")
		uuid, imgBase64, err := logic.NewImgCaptchaLogic(r.Context()).Generate(phone)
		if err != nil {
			writeCaptchaError(w, err)
			return
		}
		writeSuccess(w, types.ImgCaptchaResponse{UUID: uuid, ImgBase64: imgBase64})
	}
}

func VerifyImgCaptchaHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VerifyImgCaptchaRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if err := logic.NewImgCaptchaLogic(r.Context()).Verify(req.Phone, req.UUID, req.UserInputCode); err != nil {
			writeCaptchaError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	}
}

func InvalidateImgCaptchaHandler(_ *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.InvalidateImgCaptchaRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		if err := logic.NewImgCaptchaLogic(r.Context()).Invalidate(req.Phone, req.UUID); err != nil {
			writeCaptchaError(w, err)
			return
		}
		writeSuccess(w, types.CommonResponse{Message: "ok"})
	}
}

func writeCaptchaError(w http.ResponseWriter, err error) {
	if errors.Is(err, logic.ErrImgCaptchaInvalid) || errors.Is(err, logic.ErrImgCaptchaExpired) {
		writeError(w, http.StatusBadRequest, 41002, err.Error())
		return
	}
	writeParamError(w, err)
}
