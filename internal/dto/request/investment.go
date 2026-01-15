package request

// SyncInvestmentsRequest is the request body for syncing investments from blockchain
type SyncInvestmentsRequest struct {
	TxHash string `json:"tx_hash,omitempty"` // Optional: recent tx hash for verification
}

// WaterCropRequest is the request body for watering a crop
// No body fields needed - uses water points from user
type WaterCropRequest struct{}

// ListCropsRequest contains query parameters for listing crops
type ListCropsRequest struct {
	Status    string `form:"status"` // growing, ready, harvested
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy    string `form:"sort_by"`    // invested_at, progress, status
	SortOrder string `form:"sort_order"` // asc, desc
}

// SyncHarvestRequest is the request body for syncing harvest status
type SyncHarvestRequest struct {
	TxHash string `json:"tx_hash,omitempty"` // Optional: harvest tx hash
}
