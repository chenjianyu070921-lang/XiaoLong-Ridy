package types

// WalletTransaction 钱包流水明细，金额单位为元。
type WalletTransaction struct {
	ID        uint64  `json:"id"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	OrderID   uint64  `json:"orderId"`
	Title     string  `json:"title"`
	CreatedAt int64   `json:"createdAt"`
}

// WalletResponse 钱包余额和流水列表。
type WalletResponse struct {
	UserID       uint64              `json:"userId"`
	Balance      float64             `json:"balance"`
	Transactions []WalletTransaction `json:"transactions"`
	Total        int64               `json:"total"`
}

// WalletChangeRequest 充值或提现请求，金额单位为元。
type WalletChangeRequest struct {
	Amount float64 `json:"amount"`
}

// WalletChangeResponse 钱包变更后的余额和新增流水。
type WalletChangeResponse struct {
	Success     bool              `json:"success"`
	Balance     float64           `json:"balance"`
	Transaction WalletTransaction `json:"transaction"`
}
