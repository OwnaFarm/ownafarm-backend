package request

// AdminWalletLoginRequest represents the request body for admin wallet login
type AdminWalletLoginRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Nonce         string `json:"nonce" binding:"required"`
}

// RejectFarmerRequest represents the request body for rejecting a farmer
type RejectFarmerRequest struct {
	Reason *string `json:"reason" binding:"omitempty,max=500"`
}
