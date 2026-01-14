package repositories

import (
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// LeaderboardEntry represents a single entry in a leaderboard
type LeaderboardEntry struct {
	UserID        string          `json:"user_id"`
	WalletAddress string          `json:"wallet_address"`
	Score         decimal.Decimal `json:"score"`
	Rank          int             `json:"rank"`
}

// LeaderboardRepository defines the interface for leaderboard data access
type LeaderboardRepository interface {
	GetXPLeaderboard(limit int) ([]LeaderboardEntry, error)
	GetWealthLeaderboard(limit int) ([]LeaderboardEntry, error)
	GetProfitLeaderboard(limit int) ([]LeaderboardEntry, error)
	GetUserXPRank(userID string) (*LeaderboardEntry, error)
	GetUserWealthRank(userID string) (*LeaderboardEntry, error)
	GetUserProfitRank(userID string) (*LeaderboardEntry, error)
}

type leaderboardRepository struct {
	db *gorm.DB
}

// NewLeaderboardRepository creates a new LeaderboardRepository instance
func NewLeaderboardRepository(db *gorm.DB) LeaderboardRepository {
	return &leaderboardRepository{db: db}
}

// GetXPLeaderboard returns top users by XP
func (r *leaderboardRepository) GetXPLeaderboard(limit int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	err := r.db.Model(&models.User{}).
		Select("id as user_id, wallet_address, xp as score").
		Order("xp DESC").
		Limit(limit).
		Scan(&entries).Error

	if err != nil {
		return nil, err
	}

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
		entries[i].Score = decimal.NewFromInt(int64(entries[i].Score.IntPart()))
	}

	return entries, nil
}

// GetWealthLeaderboard returns top users by total investment amount
func (r *leaderboardRepository) GetWealthLeaderboard(limit int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	err := r.db.Model(&models.Investment{}).
		Select("investments.user_id, users.wallet_address, SUM(investments.amount) as score").
		Joins("JOIN users ON users.id = investments.user_id").
		Group("investments.user_id, users.wallet_address").
		Order("score DESC").
		Limit(limit).
		Scan(&entries).Error

	if err != nil {
		return nil, err
	}

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

// GetProfitLeaderboard returns top users by profit (harvest_amount - amount for harvested crops)
func (r *leaderboardRepository) GetProfitLeaderboard(limit int) ([]LeaderboardEntry, error) {
	var entries []LeaderboardEntry

	err := r.db.Model(&models.Investment{}).
		Select("investments.user_id, users.wallet_address, SUM(COALESCE(investments.harvest_amount, 0) - investments.amount) as score").
		Joins("JOIN users ON users.id = investments.user_id").
		Where("investments.is_harvested = ?", true).
		Group("investments.user_id, users.wallet_address").
		Order("score DESC").
		Limit(limit).
		Scan(&entries).Error

	if err != nil {
		return nil, err
	}

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

// GetUserXPRank returns the user's rank in XP leaderboard
func (r *leaderboardRepository) GetUserXPRank(userID string) (*LeaderboardEntry, error) {
	var entry LeaderboardEntry

	// Subquery to get user's rank
	subQuery := `
		SELECT user_id, wallet_address, xp as score, rank
		FROM (
			SELECT id as user_id, wallet_address, xp, 
				   RANK() OVER (ORDER BY xp DESC) as rank
			FROM users
		) ranked
		WHERE user_id = ?
	`

	err := r.db.Raw(subQuery, userID).Scan(&entry).Error
	if err != nil {
		return nil, err
	}

	if entry.UserID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &entry, nil
}

// GetUserWealthRank returns the user's rank in wealth leaderboard
func (r *leaderboardRepository) GetUserWealthRank(userID string) (*LeaderboardEntry, error) {
	var entry LeaderboardEntry

	subQuery := `
		SELECT user_id, wallet_address, score, rank
		FROM (
			SELECT i.user_id, u.wallet_address, SUM(i.amount) as score,
				   RANK() OVER (ORDER BY SUM(i.amount) DESC) as rank
			FROM investments i
			JOIN users u ON u.id = i.user_id
			GROUP BY i.user_id, u.wallet_address
		) ranked
		WHERE user_id = ?
	`

	err := r.db.Raw(subQuery, userID).Scan(&entry).Error
	if err != nil {
		return nil, err
	}

	if entry.UserID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &entry, nil
}

// GetUserProfitRank returns the user's rank in profit leaderboard
func (r *leaderboardRepository) GetUserProfitRank(userID string) (*LeaderboardEntry, error) {
	var entry LeaderboardEntry

	subQuery := `
		SELECT user_id, wallet_address, score, rank
		FROM (
			SELECT i.user_id, u.wallet_address, 
				   SUM(COALESCE(i.harvest_amount, 0) - i.amount) as score,
				   RANK() OVER (ORDER BY SUM(COALESCE(i.harvest_amount, 0) - i.amount) DESC) as rank
			FROM investments i
			JOIN users u ON u.id = i.user_id
			WHERE i.is_harvested = true
			GROUP BY i.user_id, u.wallet_address
		) ranked
		WHERE user_id = ?
	`

	err := r.db.Raw(subQuery, userID).Scan(&entry).Error
	if err != nil {
		return nil, err
	}

	if entry.UserID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &entry, nil
}
