package main

import "testing"

// TestNormalizeMySQLDSN 验证 DSN 参数补全逻辑。
// 只补齐缺失参数，不重复追加已有参数，空串保持原样。
func TestNormalizeMySQLDSN(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
		want string
	}{
		{name: "空串原样返回", dsn: "", want: ""},
		{name: "无参数时补齐三项", dsn: "root:pwd@tcp(127.0.0.1:3306)/db",
			want: "root:pwd@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"},
		{name: "已带部分参数时用 & 追加", dsn: "root:pwd@tcp(127.0.0.1:3306)/db?charset=utf8mb4",
			want: "root:pwd@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"},
		{name: "参数齐全时不重复追加", dsn: "root:pwd@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local",
			want: "root:pwd@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeMySQLDSN(tc.dsn); got != tc.want {
				t.Fatalf("normalizeMySQLDSN(%q) = %q, want %q", tc.dsn, got, tc.want)
			}
		})
	}
}
