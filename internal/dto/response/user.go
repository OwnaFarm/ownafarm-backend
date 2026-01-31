package response

import "time"

// WaterBalanceResponse represents the water balance of an investor
type WaterBalanceResponse struct {
	WaterPoints    int        `json:"water_points"`               // Current water points
	MaxWaterPoints int        `json:"max_water_points"`           // Maximum water points (100)
	NextRegenAt    *time.Time `json:"next_regen_at,omitempty"`    // When next water point regenerates (nil if at max)
	RegenRate      int        `json:"regen_rate"`                 // Minutes per 1 water point
}
