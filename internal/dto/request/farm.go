package request

// CreateFarmRequest is the request body for creating a farm
type CreateFarmRequest struct {
	Name        string   `json:"name" binding:"required,max=200"`
	Description *string  `json:"description,omitempty"`
	Location    string   `json:"location" binding:"required,max=255"`
	Latitude    *float64 `json:"latitude,omitempty" binding:"omitempty,min=-90,max=90"`
	Longitude   *float64 `json:"longitude,omitempty" binding:"omitempty,min=-180,max=180"`
	LandArea    *float64 `json:"land_area,omitempty" binding:"omitempty,min=0"` // in hectares
}

// UpdateFarmRequest is the request body for updating a farm
type UpdateFarmRequest struct {
	Name        *string  `json:"name,omitempty" binding:"omitempty,max=200"`
	Description *string  `json:"description,omitempty"`
	Location    *string  `json:"location,omitempty" binding:"omitempty,max=255"`
	Latitude    *float64 `json:"latitude,omitempty" binding:"omitempty,min=-90,max=90"`
	Longitude   *float64 `json:"longitude,omitempty" binding:"omitempty,min=-180,max=180"`
	LandArea    *float64 `json:"land_area,omitempty" binding:"omitempty,min=0"`
	CCTVUrl     *string  `json:"cctv_url,omitempty"`
	CCTVImageUrl *string `json:"cctv_image_url,omitempty"`
}

// ListFarmRequest contains query parameters for listing farms
type ListFarmRequest struct {
	IsActive  *bool  `form:"is_active"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	SortBy    string `form:"sort_by"`    // created_at, name, location, land_area
	SortOrder string `form:"sort_order"` // asc, desc
	Search    string `form:"search"`     // Search in name, location
}
