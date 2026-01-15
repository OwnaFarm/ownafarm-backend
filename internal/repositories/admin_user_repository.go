package repositories

import (
	"context"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/models"
	"gorm.io/gorm"
)

type AdminUserRepository interface {
	GetByWalletAddress(ctx context.Context, walletAddress string) (*models.AdminUser, error)
	UpdateLastLogin(ctx context.Context, adminID string) error
}

type adminUserRepository struct {
	db *gorm.DB
}

func NewAdminUserRepository(db *gorm.DB) AdminUserRepository {
	return &adminUserRepository{db: db}
}

// GetByWalletAddress retrieves an admin user by their wallet address
func (r *adminUserRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.AdminUser, error) {
	var admin models.AdminUser
	err := r.db.WithContext(ctx).Where("wallet_address = ?", walletAddress).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLogin updates the last_login_at timestamp for an admin user
func (r *adminUserRepository) UpdateLastLogin(ctx context.Context, adminID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.AdminUser{}).
		Where("id = ?", adminID).
		Update("last_login_at", &now).Error
}
