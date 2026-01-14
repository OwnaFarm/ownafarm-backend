package response

import (
	"time"

	"github.com/shopspring/decimal"
)

// FarmResponse is the response for a single farm
type FarmResponse struct {
	ID              string           `json:"id"`
	FarmerID        string           `json:"farmer_id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description,omitempty"`
	Location        string           `json:"location"`
	Latitude        *decimal.Decimal `json:"latitude,omitempty"`
	Longitude       *decimal.Decimal `json:"longitude,omitempty"`
	LandArea        *decimal.Decimal `json:"land_area,omitempty"`
	CCTVUrl         *string          `json:"cctv_url,omitempty"`
	CCTVImageUrl    *string          `json:"cctv_image_url,omitempty"`
	CCTVLastUpdated *time.Time       `json:"cctv_last_updated,omitempty"`
	IsActive        bool             `json:"is_active"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// FarmListItem represents a farm in the list response
type FarmListItem struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Location  string           `json:"location"`
	LandArea  *decimal.Decimal `json:"land_area,omitempty"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
}

// ListFarmResponse is the response for listing farms
type ListFarmResponse struct {
	Farms      []FarmListItem `json:"farms"`
	Pagination PaginationMeta `json:"pagination"`
}

// CreateFarmResponse is the response for creating a farm
type CreateFarmResponse struct {
	Farm FarmResponse `json:"farm"`
}
