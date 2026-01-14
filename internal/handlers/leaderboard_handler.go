package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/dto/request"
	"github.com/ownafarm/ownafarm-backend/internal/services"
)

// LeaderboardHandler handles leaderboard-related HTTP requests
type LeaderboardHandler struct {
	leaderboardService services.LeaderboardServiceInterface
}

// NewLeaderboardHandler creates a new LeaderboardHandler instance
func NewLeaderboardHandler(leaderboardService services.LeaderboardServiceInterface) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

// GetLeaderboard retrieves the leaderboard
// GET /leaderboard?type=<xp|wealth|profit>&limit=<N>
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req request.GetLeaderboardRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default limit to 10 if not provided
	if req.Limit <= 0 {
		req.Limit = 10
	}

	resp, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), userID.(string), req.Type, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
