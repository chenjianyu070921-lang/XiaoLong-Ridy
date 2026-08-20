package channel

import (
	"os"
	"path/filepath"
	"testing"

	commonalipay "XiaoLong-Ridy/common/alipay"
)

// TestAlipayChannel_FromYamlConfig 验证从实际 paysvc.yaml 读取配置能成功构造支付宝渠道。
// 该测试读取真实配置文件，验证沙箱密钥可被 SDK 正确解析。
func TestAlipayChannel_FromYamlConfig(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..") // 从 internal/channel 回溯到项目根
	yamlPath := filepath.Join(root, "rpc", "paysvc", "etc", "paysvc.yaml")
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Skipf("skip: cannot read config %s: %v", yamlPath, err)
	}

	// 简单解析 appId/privateKey/alipayPublicKey（依赖 go-zero conf 读取）
	// 这里用 YAML 手动校验，仅做 smoke 验证
	cfg := commonalipay.Config{
		AppId:           findYaml(data, "appId"),
		PrivateKey:      findYaml(data, "privateKey"),
		AlipayPublicKey: findYaml(data, "alipayPublicKey"),
		Sandbox:         true,
	}
	if cfg.AppId == "" {
		t.Skip("skip: appId empty in config")
	}

	ch, err := NewAlipayChannel(cfg)
	if err != nil {
		t.Fatalf("NewAlipayChannel failed with sandbox keys: %v", err)
	}
	if ch == nil || ch.Name() != "alipay" {
		t.Fatalf("unexpected channel: %+v", ch)
	}
	t.Log("alipay channel init OK with sandbox keys")
}

// findYaml 简易提取 yaml 顶层键值（仅本测试用）。
func findYaml(data []byte, key string) string {
	prefix := key + ":"
	for _, line := range splitLines(string(data)) {
		line = trim(line)
		if len(line) > len(prefix) && line[:len(prefix)] == prefix {
			v := trim(line[len(prefix):])
			// 去掉单/双引号
			if len(v) >= 2 && (v[0] == '\'' || v[0] == '"') && v[len(v)-1] == v[0] {
				v = v[1 : len(v)-1]
			}
			return v
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
