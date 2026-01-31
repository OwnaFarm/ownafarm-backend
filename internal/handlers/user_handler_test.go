package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockUserRepository struct {
	GetByIDFunc            func(id string) (*models.User, error)
	GetByWalletAddressFunc func(walletAddress string) (*models.User, error)
	CreateFunc             func(user *models.User) error
	UpdateLastLoginFunc    func(userID string) error
	RegenerateWaterFunc    func(userID string) (*models.User, error)
	UpdateGameStatsFunc    func(userID string, updates map[string]interface{}) error
}

func (m *mockUserRepository) UpdateGameStats(userID string, updates map[string]interface{}) error {
	if m.UpdateGameStatsFunc != nil {
		return m.UpdateGameStatsFunc(userID, updates)
	}
	return nil
}

func (m *mockUserRepository) RegenerateWater(userID string) (*models.User, error) {
	if m.RegenerateWaterFunc != nil {
		return m.RegenerateWaterFunc(userID)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByID(id string) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *mockUserRepository) GetByWalletAddress(walletAddress string) (*models.User, error) {
	if m.GetByWalletAddressFunc != nil {
		return m.GetByWalletAddressFunc(walletAddress)
	}
	return nil, nil
}

func (m *mockUserRepository) Create(user *models.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return nil
}

func (m *mockUserRepository) UpdateLastLogin(userID string) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(userID)
	}
	return nil
}

func TestGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			GetByIDFunc: func(id string) (*models.User, error) {
				name := "John Doe"
				return &models.User{
					ID:   id,
					Name: &name,
				}, nil
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/:id", handler.GetByID)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/550e8400-e29b-41d4-a716-446655440000", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"success"`)
		assert.Contains(t, w.Body.String(), `"id":"550e8400-e29b-41d4-a716-446655440000"`)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		handler := NewUserHandler(&mockUserRepository{})
		router := gin.New()
		router.GET("/users/:id", handler.GetByID)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/invalid-uuid", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid user ID format")
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			GetByIDFunc: func(id string) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/:id", handler.GetByID)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/550e8400-e29b-41d4-a716-446655440000", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			GetByIDFunc: func(id string) (*models.User, error) {
				return nil, errors.New("db error")
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/:id", handler.GetByID)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/550e8400-e29b-41d4-a716-446655440000", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
	})
}

func TestGetWaterBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		mockRepo := &mockUserRepository{
			RegenerateWaterFunc: func(userID string) (*models.User, error) {
				return &models.User{
					ID:          userID,
					WaterPoints: 75,
					LastRegenAt: &now,
				}, nil
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/me/water", func(c *gin.Context) {
			c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
			handler.GetWaterBalance(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me/water", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"success"`)
		assert.Contains(t, w.Body.String(), `"water_points":75`)
		assert.Contains(t, w.Body.String(), `"max_water_points":100`)
		assert.Contains(t, w.Body.String(), `"regen_rate":5`)
	})

	t.Run("unauthorized - missing user_id", func(t *testing.T) {
		handler := NewUserHandler(&mockUserRepository{})
		router := gin.New()
		router.GET("/users/me/water", handler.GetWaterBalance)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me/water", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			RegenerateWaterFunc: func(userID string) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/me/water", func(c *gin.Context) {
			c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
			handler.GetWaterBalance(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me/water", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("internal server error", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			RegenerateWaterFunc: func(userID string) (*models.User, error) {
				return nil, errors.New("db error")
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/me/water", func(c *gin.Context) {
			c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
			handler.GetWaterBalance(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me/water", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
	})

	t.Run("at max water points - no next_regen_at", func(t *testing.T) {
		now := time.Now()
		mockRepo := &mockUserRepository{
			RegenerateWaterFunc: func(userID string) (*models.User, error) {
				return &models.User{
					ID:          userID,
					WaterPoints: 100, // At max
					LastRegenAt: &now,
				}, nil
			},
		}
		handler := NewUserHandler(mockRepo)
		router := gin.New()
		router.GET("/users/me/water", func(c *gin.Context) {
			c.Set("user_id", "550e8400-e29b-41d4-a716-446655440000")
			handler.GetWaterBalance(c)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/me/water", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"water_points":100`)
		// next_regen_at should be null when at max
		assert.NotContains(t, w.Body.String(), `"next_regen_at":"`)
	})
}

