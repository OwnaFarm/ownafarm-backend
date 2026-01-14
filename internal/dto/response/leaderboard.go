package response

// LeaderboardEntryResponse represents a single entry in a leaderboard
type LeaderboardEntryResponse struct {
	Rank          int     `json:"rank"`
	WalletAddress string  `json:"wallet_address"`
	Score         float64 `json:"score"`
	IsCurrentUser bool    `json:"is_current_user"`
}

// LeaderboardResponse represents the leaderboard response
type LeaderboardResponse struct {
	Type      string                     `json:"type"` // xp, wealth, profit
	Entries   []LeaderboardEntryResponse `json:"entries"`
	UserEntry *LeaderboardEntryResponse  `json:"user_entry,omitempty"` // Current user's position if not in entries
}
