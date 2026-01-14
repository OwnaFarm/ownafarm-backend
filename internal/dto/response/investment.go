package response

import "time"

// CropResponse represents a single crop/investment for display
type CropResponse struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`                     // From invoice
	Image         *string  `json:"image"`                    // From invoice
	CCTVImage     *string  `json:"cctv_image"`               // From farm
	Location      string   `json:"location"`                 // From farm
	Progress      int      `json:"progress"`                 // 0-100
	DaysLeft      int      `json:"days_left"`                // Remaining days until harvest
	YieldPercent  float64  `json:"yield_percent"`            // Return percentage
	Invested      float64  `json:"invested"`                 // GOLD amount invested
	Status        string   `json:"status"`                   // growing, ready, harvested
	PlantedAt     string   `json:"planted_at"`               // ISO timestamp
	WaterCount    int      `json:"water_count"`              // Times watered
	CanHarvest    bool     `json:"can_harvest"`              // Is mature & not harvested
	HarvestAmount *float64 `json:"harvest_amount,omitempty"` // Amount received after harvest
}

// SyncInvestmentsResponse represents the response for syncing investments
type SyncInvestmentsResponse struct {
	SyncedCount int            `json:"synced_count"` // Number of new investments synced
	NewCrops    []CropResponse `json:"new_crops"`    // Newly synced crops
}

// ListCropsResponse represents the response for listing crops
type ListCropsResponse struct {
	Crops      []CropResponse `json:"crops"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

// WaterCropResponse represents the response after watering a crop
type WaterCropResponse struct {
	Crop           CropResponse `json:"crop"`
	XPGained       int          `json:"xp_gained"`
	WaterRemaining int          `json:"water_remaining"` // User's remaining water points
}

// SyncHarvestResponse represents the response after syncing harvest status
type SyncHarvestResponse struct {
	Crop     CropResponse `json:"crop"`
	XPGained int          `json:"xp_gained"`
}

// UserGameStats represents user's game statistics (for sync response)
type UserGameStats struct {
	Level       int        `json:"level"`
	XP          int        `json:"xp"`
	WaterPoints int        `json:"water_points"`
	LastRegenAt *time.Time `json:"last_regen_at,omitempty"`
}
