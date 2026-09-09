package types

import (
	"encoding/json"
	"testing"
)

func TestRegisterDriverRequestAcceptsLegacySnakeCaseFields(t *testing.T) {
	var request RegisterDriverRequest
	err := json.Unmarshal([]byte(`{
		"phone":"15789654687",
		"password_hash":"852963741456789",
		"real_name":" 张三 ",
		"id_card_no":"11010119900101123x",
		"driver_license_no":"DL10000001",
		"avatar_url":"avatar"
	}`), &request)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if request.Phone != "15789654687" || request.Password != "852963741456789" || request.RealName != " 张三 " || request.IdCardNo != "11010119900101123x" || request.DriverLicenseNo != "DL10000001" || request.AvatarURL != "avatar" {
		t.Fatalf("unexpected request: %#v", request)
	}
}

func TestRegisterDriverRequestPrefersCanonicalFields(t *testing.T) {
	var request RegisterDriverRequest
	err := json.Unmarshal([]byte(`{
		"password":"canonical-password",
		"password_hash":"legacy-password",
		"realName":"标准姓名",
		"real_name":"旧姓名"
	}`), &request)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if request.Password != "canonical-password" || request.RealName != "标准姓名" {
		t.Fatalf("unexpected request: %#v", request)
	}
}
