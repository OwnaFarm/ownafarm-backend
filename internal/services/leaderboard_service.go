package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/dto/response"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/valkey-io/valkey-go"
)

const (
	// LeaderboardCacheTTL is the TTL for leaderboard cache
	LeaderboardCacheTTL = 5 * time.Minute
)

// LeaderboardServiceInterface defines the interface for leaderboard operations
type LeaderboardServiceInterface interface {
	GetLeaderboard(ctx context.Context, userID string, leaderboardType string, limit int) (*response.LeaderboardResponse, error)
}

// LeaderboardService handles leaderboard business logic
type LeaderboardService struct {
	repo     repositories.LeaderboardRepository
	valkey   valkey.Client
	cacheTTL time.Duration
}

// NewLeaderboardService creates a new LeaderboardService instance
func NewLeaderboardService(repo repositories.LeaderboardRepository, valkeyClient valkey.Client) *LeaderboardService {
	return &LeaderboardService{
		repo:     repo,
		valkey:   valkeyClient,
		cacheTTL: LeaderboardCacheTTL,
	}
}

// GetLeaderboard retrieves the leaderboard for a given type
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, userID string, leaderboardType string, limit int) (*response.LeaderboardResponse, error) {
	// Default limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Try to get from cache
	cacheKey := s.cacheKey(leaderboardType, limit)
	entries, err := s.getFromCache(ctx, cacheKey)
	if err != nil || entries == nil {
		// Cache miss - fetch from database
		entries, err = s.fetchFromDB(leaderboardType, limit)
		if err != nil {
			return nil, err
		}

		// Store in cache
		_ = s.setCache(ctx, cacheKey, entries)
	}

	// Convert to response
	responseEntries := make([]response.LeaderboardEntryResponse, len(entries))
	userInTopN := false

	for i, entry := range entries {
		isCurrentUser := entry.UserID == userID
		if isCurrentUser {
			userInTopN = true
		}

		responseEntries[i] = response.LeaderboardEntryResponse{
			Rank:          entry.Rank,
			WalletAddress: entry.WalletAddress,
			Score:         entry.Score.InexactFloat64(),
			IsCurrentUser: isCurrentUser,
		}
	}

	resp := &response.LeaderboardResponse{
		Type:    leaderboardType,
		Entries: responseEntries,
	}

	// If user is not in top-N, get their rank separately (not cached)
	if !userInTopN {
		userEntry, err := s.getUserRank(userID, leaderboardType)
		if err == nil && userEntry != nil {
			resp.UserEntry = &response.LeaderboardEntryResponse{
				Rank:          userEntry.Rank,
				WalletAddress: userEntry.WalletAddress,
				Score:         userEntry.Score.InexactFloat64(),
				IsCurrentUser: true,
			}
		}
		// If error, just skip user entry (user might have no investments/profit)
	}

	return resp, nil
}

// cacheKey generates cache key for leaderboard
func (s *LeaderboardService) cacheKey(leaderboardType string, limit int) string {
	return fmt.Sprintf("leaderboard:%s:%d", leaderboardType, limit)
}

// getFromCache retrieves leaderboard from Valkey cache
func (s *LeaderboardService) getFromCache(ctx context.Context, key string) ([]repositories.LeaderboardEntry, error) {
	cmd := s.valkey.B().Get().Key(key).Build()
	data, err := s.valkey.Do(ctx, cmd).ToString()
	if err != nil {
		return nil, err
	}

	var entries []repositories.LeaderboardEntry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

// setCache stores leaderboard in Valkey cache
func (s *LeaderboardService) setCache(ctx context.Context, key string, entries []repositories.LeaderboardEntry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	cmd := s.valkey.B().Set().Key(key).Value(string(data)).Ex(s.cacheTTL).Build()
	return s.valkey.Do(ctx, cmd).Error()
}

// fetchFromDB fetches leaderboard from database
func (s *LeaderboardService) fetchFromDB(leaderboardType string, limit int) ([]repositories.LeaderboardEntry, error) {
	switch leaderboardType {
	case "xp":
		return s.repo.GetXPLeaderboard(limit)
	case "wealth":
		return s.repo.GetWealthLeaderboard(limit)
	case "profit":
		return s.repo.GetProfitLeaderboard(limit)
	default:
		return nil, fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}
}

// getUserRank gets user's rank for a specific leaderboard type
func (s *LeaderboardService) getUserRank(userID, leaderboardType string) (*repositories.LeaderboardEntry, error) {
	switch leaderboardType {
	case "xp":
		return s.repo.GetUserXPRank(userID)
	case "wealth":
		return s.repo.GetUserWealthRank(userID)
	case "profit":
		return s.repo.GetUserProfitRank(userID)
	default:
		return nil, fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}
}
