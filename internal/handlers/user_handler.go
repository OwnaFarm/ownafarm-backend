package handlers

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"gorm.io/gorm"
)

type UserHandler struct {
	repo repositories.UserRepository
}

func NewUserHandler(repo repositories.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

var uuidRegex = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	// Validate UUID format
	if !uuidRegex.MatchString(id) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID format",
		})
		return
	}

	user, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status":  "error",
				"message": "User not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}
