package sensitive

import "strings"

// keywords 命中的关键词一律拦截发送，防止司乘绕开平台私下交易/留联系方式。
// 覆盖方案文档要求：手机号/微信号/QQ/支付宝/线下交易 等。
var keywords = []string{
	"微信", "微信号", "wx", "WX", "WeChat", "wechat",
	"qq", "QQ", "扣扣",
	"支付宝", "支付宝账号", "zfb", "ZFB",
	"手机号", "电话号", "联系电话", "加我",
	"私下", "线下", "当面", "现金", "转账", "红包", "私聊", "私信",
	"站外", "平台外", "加好友",
}

// Contains 判断消息内容是否命中敏感词。
func Contains(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	for _, kw := range keywords {
		if strings.Contains(content, kw) {
			return true
		}
	}
	return false
}
