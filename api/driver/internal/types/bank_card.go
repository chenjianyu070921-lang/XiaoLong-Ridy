package types

// ---- 银行卡与平台提现密码 ----
// 合规红线：绑卡全程不采集银行卡取款密码；银行卡号接口只返回脱敏视图。

// SendBankCardSmsCodeRequest 绑卡短信验证码：发往银行预留手机号。
type SendBankCardSmsCodeRequest struct {
	ReservedPhone string `json:"reservedPhone"`
}

// BindBankCardRequest 绑卡请求。
type BindBankCardRequest struct {
	BankName     string `json:"bankName"`
	CardNo       string `json:"cardNo"`
	HolderName   string `json:"holderName"`
	HolderIdCard string `json:"holderIdCard"`
	ReservedPhone string `json:"reservedPhone"`
	SmsCode      string `json:"smsCode"`
}

// BindBankCardResponse 绑卡结果。
type BindBankCardResponse struct {
	ID         int64  `json:"id"`
	BankName   string `json:"bankName"`
	MaskedCardNo string `json:"maskedCardNo"`
}

// BankCardInfo 银行卡脱敏视图。
type BankCardInfo struct {
	ID           int64  `json:"id"`
	BankName     string `json:"bankName"`
	MaskedCardNo string `json:"maskedCardNo"`
	HolderName   string `json:"holderName"`
	ReservedPhone string `json:"reservedPhone"`
	CreatedAt    int64  `json:"createdAt"`
}

// ListBankCardsResponse 银行卡列表。
type ListBankCardsResponse struct {
	Cards []BankCardInfo `json:"cards"`
}

// DeleteBankCardRequest 删除银行卡。
type DeleteBankCardRequest struct {
	ID int64 `json:"id"`
}

// ResetWithdrawPasswordRequest 遗忘提现密码找回：三重实名校验。
type ResetWithdrawPasswordRequest struct {
	Phone    string `json:"phone"`
	IdCardNo string `json:"idCardNo"`
	RealName string `json:"realName"`
}

// ResetWithdrawPasswordResponse 找回结果（新密码仅通过短信下发，不返回明文）。
type ResetWithdrawPasswordResponse struct {
	Message string `json:"message"`
}
