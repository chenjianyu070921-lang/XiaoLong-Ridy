package logic

import (
	"testing"

	"XiaoLong-Ridy/api/driver/internal/types"
)

func TestValidPhone(t *testing.T) {
	cases := []struct {
		phone string
		want  bool
	}{
		{"13800000001", true},   // 常规 11 位手机号
		{"19912345678", true},   // 19 开头（第二位 9）
		{"14712345678", true},   // 14 开头（第二位 4）
		{"12800000001", false},  // 第二位 2 非法
		{"1380000000", false},   // 10 位
		{"138000000011", false}, // 12 位
		{"", false},
		{"1380000000a", false}, // 含字母
	}
	for _, c := range cases {
		if got := validPhone(c.phone); got != c.want {
			t.Errorf("validPhone(%q) = %v, want %v", c.phone, got, c.want)
		}
	}
}

func TestValidIDCard(t *testing.T) {
	cases := []struct {
		no   string
		want bool
	}{
		{"110101199001011234", true},   // 18 位数字
		{"11010119900101123X", true},   // 末位大写 X
		{"11010119900101123x", true},   // 末位小写 x
		{"1101011990010112", false},    // 17 位
		{"1101011990010112345", false}, // 19 位
		{"11010119900101123Y", false},  // 末位非法字符
		{"11010119900101123", false},   // 末位缺失
		{"", false},
	}
	for _, c := range cases {
		if got := validIDCard(c.no); got != c.want {
			t.Errorf("validIDCard(%q) = %v, want %v", c.no, got, c.want)
		}
	}
}

func TestValidPassword(t *testing.T) {
	cases := []struct {
		password string
		want     bool
	}{
		{"1234567", false},
		{"12345678", true},
		{"Driver@123", true},
		{"", false},
		{string(make([]byte, 73)), false},
	}
	for _, testCase := range cases {
		if got := validPassword(testCase.password); got != testCase.want {
			t.Errorf("validPassword(%q) = %v, want %v", testCase.password, got, testCase.want)
		}
	}
}

func TestNormalizeRegisterDriverRequest(t *testing.T) {
	req := &types.RegisterDriverRequest{
		Phone:           " 13800000001 ",
		RealName:        " 张三 ",
		IdCardNo:        " 11010119900101123x ",
		DriverLicenseNo: " DL10000001 ",
		AvatarURL:       " avatar ",
	}
	normalizeRegisterDriverRequest(req)
	if req.Phone != "13800000001" || req.RealName != "张三" || req.IdCardNo != "11010119900101123X" || req.DriverLicenseNo != "DL10000001" || req.AvatarURL != "avatar" {
		t.Fatalf("unexpected normalized request: %#v", req)
	}
}

func TestValidDriverStatus(t *testing.T) {
	if !validDriverStatus("DRIVER_STATUS_NORMAL") {
		t.Fatal("validDriverStatus() rejected a valid status")
	}
	if validDriverStatus("DRIVER_STATUS_UNKNOWN") {
		t.Fatal("validDriverStatus() accepted an unknown status")
	}
}

func TestClampPage(t *testing.T) {
	cases := []struct {
		page, pageSize     int32
		wantPage, wantSize int32
	}{
		{1, 20, 1, 20},       // 默认值保持
		{0, 0, 1, 20},        // 小于 1 收敛到默认
		{-3, -5, 1, 20},      // 负数收敛到默认
		{5, 200, 5, 100},     // 超过上限收敛到 100
		{5, 0, 5, 20},        // pageSize 为 0 用默认
		{100, 1, 100, 1},     // 边界：pageSize 最小 1
		{100, 101, 100, 100}, // 边界：pageSize 超上限
	}
	for _, c := range cases {
		gotPage, gotSize := clampPage(c.page, c.pageSize)
		if gotPage != c.wantPage || gotSize != c.wantSize {
			t.Errorf("clampPage(%d,%d) = (%d,%d), want (%d,%d)",
				c.page, c.pageSize, gotPage, gotSize, c.wantPage, c.wantSize)
		}
	}
}
