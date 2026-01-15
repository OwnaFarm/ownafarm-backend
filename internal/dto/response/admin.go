package response

import "time"

// AdminInfo represents the admin user information in responses
type AdminInfo struct {
	ID            string `json:"id"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
}

// AdminNonceResponse represents the response body for admin nonce request
type AdminNonceResponse struct {
	Nonce       string `json:"nonce"`
	SignMessage string `json:"sign_message"`
}

// AdminLoginResponse represents the response body for successful admin login
type AdminLoginResponse struct {
	Token string    `json:"token"`
	Admin AdminInfo `json:"admin"`
}

// AdminLoginErrorResponse represents the response body for failed admin login
type AdminLoginErrorResponse struct {
	Error             string `json:"error"`
	RemainingAttempts *int   `json:"remaining_attempts,omitempty"`
	RetryAfterSeconds *int64 `json:"retry_after_seconds,omitempty"`
}

// FarmerStatusUpdateResponse represents the response for farmer approve/reject
type FarmerStatusUpdateResponse struct {
	FarmerID   string    `json:"farmer_id"`
	Status     string    `json:"status"`
	ReviewedBy string    `json:"reviewed_by"`
	ReviewedAt time.Time `json:"reviewed_at"`
	Reason     *string   `json:"reason,omitempty"`
}
