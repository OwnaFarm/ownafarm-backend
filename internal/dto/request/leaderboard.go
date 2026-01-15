package request

// GetLeaderboardRequest contains query parameters for leaderboard
type GetLeaderboardRequest struct {
	Type  string `form:"type" binding:"required,oneof=xp wealth profit"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
}
