// Package alipay 定义支付宝支付渠道的配置结构，供各服务 channel 层引用。
package alipay

// Config 支付宝支付配置。
//
// 各服务在自己的 config 中嵌入：
//
//	Alipay alipay.Config `yaml:"alipay" json:"alipay"`
type Config struct {
	// AppId 应用ID（支付宝开放平台分配）
	AppId string `yaml:"appId" json:"appId"`
	// PrivateKey 应用私钥（用于请求签名）
	PrivateKey string `yaml:"privateKey" json:"privateKey"`
	// AlipayPublicKey 支付宝公钥（用于验签回调）
	AlipayPublicKey string `yaml:"alipayPublicKey" json:"alipayPublicKey"`
	// Gateway 支付宝网关地址，默认生产环境
	Gateway string `yaml:"gateway" json:"gateway"`
	// NotifyUrl 支付结果异步回调地址
	NotifyUrl string `yaml:"notifyUrl" json:"notifyUrl"`
	// ReturnUrl 支付完成同步跳转地址（可选，App 支付可留空）
	ReturnUrl string `yaml:"returnUrl" json:"returnUrl"`
	// SignType 签名算法：RSA2（推荐）/ RSA
	SignType string `yaml:"signType" json:"signType"`
	// Charset 字符编码，固定 utf-8
	Charset string `yaml:"charset" json:"charset"`
	// TimeoutExpress 交易超时时间，如 "30m" 表示 30 分钟
	TimeoutExpress string `yaml:"timeoutExpress" json:"timeoutExpress"`
	// Sandbox 是否沙箱环境（调试用）
	Sandbox bool `yaml:"sandbox" json:"sandbox"`
}

// 支付宝相关常量
const (
	// GatewayProduction 支付宝生产环境网关
	GatewayProduction = "https://openapi.alipay.com/gateway.do"
	// GatewaySandbox 支付宝沙箱环境网关
	GatewaySandbox = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"

	// SignTypeRSA2 RSA2 签名（推荐）
	SignTypeRSA2 = "RSA2"
	// SignTypeRSA RSA 签名
	SignTypeRSA = "RSA"

	// CharsetUTF8 字符编码
	CharsetUTF8 = "utf-8"

	// FormatJSON 请求/响应格式
	FormatJSON = "JSON"
)

// WithDefaults 返回填充默认值后的配置副本。
// 未显式设置的字段使用支付宝通用默认值。
func (c Config) WithDefaults() Config {
	if c.Gateway == "" {
		c.Gateway = GatewayProduction
	}
	if c.SignType == "" {
		c.SignType = SignTypeRSA2
	}
	if c.Charset == "" {
		c.Charset = CharsetUTF8
	}
	if c.Sandbox {
		c.Gateway = GatewaySandbox
	}
	return c
}
