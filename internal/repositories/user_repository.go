package repositories

import (
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetByID(id string) (*models.User, error)
	GetByWalletAddress(walletAddress string) (*models.User, error)
	Create(user *models.User) error
	UpdateLastLogin(userID string) error
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
