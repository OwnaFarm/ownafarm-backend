package repositories

import (
	"errors"
	"math"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

const (
	// WaterRegenRateMinutes is the number of minutes per 1 water point regenerated
	WaterRegenRateMinutes = 5
	// MaxWaterPoints is the maximum water points a user can have
	MaxWaterPoints = 100
)

var (
	ErrNotEnoughWater = errors.New("not enough water points")
)

type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByWalletAddress(walletAddress string) (*models.User, error)
	Create(user *models.User) error
	UpdateLastLogin(userID string) error
	UpdateGameStats(userID string, updates map[string]interface{}) error
	RegenerateWater(userID string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByWalletAddress(walletAddress string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "wallet_address = ?", walletAddress).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) UpdateLastLogin(userID string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", now).Error
}

// UpdateGameStats updates user game statistics with optimistic locking on water_points
// The updates map can contain: water_points, xp, level
// For water_points updates, it checks that current water_points >= required amount
func (r *userRepository) UpdateGameStats(userID string, updates map[string]interface{}) error {
	query := r.db.Model(&models.User{}).Where("id = ?", userID)

	// If updating water_points with a negative delta, add optimistic locking
	if waterDelta, ok := updates["water_points"]; ok {
		// Get current user to calculate new water value
		var user models.User
		if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}

		// If it's a numeric value, treat as absolute; if negative, it's a deduction
		newWater := waterDelta.(int)
		if newWater < user.WaterPoints {
			// This is a deduction, use optimistic locking
			requiredWater := user.WaterPoints - newWater
			query = query.Where("water_points >= ?", requiredWater)
		}
	}

	// Perform update
	updates["updated_at"] = time.Now()
	result := query.Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	// If no rows affected and we were checking water, it means concurrent update or insufficient water
	if result.RowsAffected == 0 {
		if _, ok := updates["water_points"]; ok {
			return ErrNotEnoughWater
		}
		return gorm.ErrRecordNotFound
	}

	// Check if level-up is needed after XP update
	if _, xpUpdated := updates["xp"]; xpUpdated {
		var user models.User
		if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}

		newLevel := calculateLevel(user.XP)
		if newLevel != user.Level {
			if err := r.db.Model(&models.User{}).Where("id = ?", userID).Update("level", newLevel).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// RegenerateWater regenerates water points based on time elapsed since last regeneration
// Returns the updated user with fresh water points
func (r *userRepository) RegenerateWater(userID string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	// If LastRegenAt is nil, set it to now and return
	if user.LastRegenAt == nil {
		now := time.Now()
		user.LastRegenAt = &now
		if err := r.db.Model(&user).Updates(map[string]interface{}{
			"last_regen_at": now,
			"updated_at":    now,
		}).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	// Calculate water to regenerate
	now := time.Now()
	elapsedMinutes := now.Sub(*user.LastRegenAt).Minutes()
	waterToAdd := int(math.Floor(elapsedMinutes / WaterRegenRateMinutes))

	// If no water to add, return current user
	if waterToAdd == 0 {
		return &user, nil
	}

	// Calculate new water points (capped at max)
	newWaterPoints := user.WaterPoints + waterToAdd
	if newWaterPoints > MaxWaterPoints {
		newWaterPoints = MaxWaterPoints
	}

	// Calculate new LastRegenAt based on complete regeneration cycles
	completeRegenCycles := int(math.Floor(elapsedMinutes / WaterRegenRateMinutes))
	newLastRegenAt := user.LastRegenAt.Add(time.Duration(completeRegenCycles*WaterRegenRateMinutes) * time.Minute)

	// Update database
	if err := r.db.Model(&user).Updates(map[string]interface{}{
		"water_points":  newWaterPoints,
		"last_regen_at": newLastRegenAt,
		"updated_at":    now,
	}).Error; err != nil {
		return nil, err
	}

	// Update user object with new values
	user.WaterPoints = newWaterPoints
	user.LastRegenAt = &newLastRegenAt
	user.UpdatedAt = now

	return &user, nil
}

// calculateLevel calculates the user's level based on XP
// Formula: level = 1 + floor(sqrt(xp/50))
// This gives smooth progression: L1->L2 at 50 XP, L2->L3 at 200 XP, L3->L4 at 450 XP, etc.
func calculateLevel(xp int) int {
	if xp < 0 {
		return 1
	}
	return 1 + int(math.Floor(math.Sqrt(float64(xp)/50.0)))
}
