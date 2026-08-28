package types

import (
	"encoding/json"
	"testing"
)

func TestVerifyImgCaptchaRequestAcceptsCanonicalCodeField(t *testing.T) {
	var req VerifyImgCaptchaRequest
	if err := json.Unmarshal([]byte(`{"phone":"13800000000","uuid":"captcha-1","code":"1234"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Code != "1234" {
		t.Fatalf("captcha code = %q, want %q", req.Code, "1234")
	}
}

func TestVerifyImgCaptchaRequestAcceptsLegacyUserInputCodeField(t *testing.T) {
	var req VerifyImgCaptchaRequest
	if err := json.Unmarshal([]byte(`{"phone":"13800000000","uuid":"captcha-1","userInputCode":"5678"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Code != "5678" {
		t.Fatalf("legacy captcha code = %q, want %q", req.Code, "5678")
	}
}

func TestVerifyImgCaptchaRequestAcceptsCaptchaAliases(t *testing.T) {
	var req VerifyImgCaptchaRequest
	if err := json.Unmarshal([]byte(`{"phone":"13800000000","captchaId":"captcha-2","captchaCode":"9012"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.UUID != "captcha-2" {
		t.Fatalf("captcha uuid = %q, want %q", req.UUID, "captcha-2")
	}
	if req.Code != "9012" {
		t.Fatalf("captcha code = %q, want %q", req.Code, "9012")
	}
}
