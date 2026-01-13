package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/models"
	"github.com/ownafarm/ownafarm-backend/internal/repositories"
	"github.com/ownafarm/ownafarm-backend/internal/utils"
	"github.com/valkey-io/valkey-go"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid wallet address or signature")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrAdminNotFound      = errors.New("admin not found")
)

// AdminLoginResult contains the result of a successful admin login
type AdminLoginResult struct {
	Token string
	Admin *models.AdminUser
}

// RateLimitInfo contains rate limit information for error responses
type RateLimitInfo struct {
	Remaining  int
	RetryAfter int64
}

// AdminNonceResult contains the result of nonce generation
type AdminNonceResult struct {
	Nonce       string
	SignMessage string
}

type AdminAuthService struct {
	adminRepo        repositories.AdminUserRepository
	rateLimitService *RateLimitService
	adminJwtUtil     *utils.AdminJWTUtil
	authService      *AuthService
	valkeyClient     valkey.Client
	nonceTTL         time.Duration
}

func NewAdminAuthService(
	adminRepo repositories.AdminUserRepository,
	rateLimitService *RateLimitService,
	adminJwtUtil *utils.AdminJWTUtil,
	authService *AuthService,
	valkeyClient valkey.Client,
	nonceTTLMinutes int,
) *AdminAuthService {
	return &AdminAuthService{
		adminRepo:        adminRepo,
		rateLimitService: rateLimitService,
		adminJwtUtil:     adminJwtUtil,
		authService:      authService,
		valkeyClient:     valkeyClient,
		nonceTTL:         time.Duration(nonceTTLMinutes) * time.Minute,
	}
}

// adminNonceKey generates the Valkey key for admin nonce (separate from user nonce)
func (s *AdminAuthService) adminNonceKey(walletAddress string) string {
	return fmt.Sprintf("admin_nonce:%s", strings.ToLower(walletAddress))
}

// generateRandomNonce generates a cryptographically secure random nonce
func (s *AdminAuthService) generateRandomNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// buildSignMessage builds the message to be signed by the admin
func (s *AdminAuthService) buildSignMessage(nonce string) string {
	return fmt.Sprintf("Sign this message to login to OwnaFarm Admin.\n\nNonce: %s", nonce)
}

// GetNonce generates and stores a nonce for admin wallet authentication
func (s *AdminAuthService) GetNonce(ctx context.Context, walletAddress string) (*AdminNonceResult, error) {
	// Normalize wallet address
	normalizedAddress := s.authService.NormalizeWalletAddress(walletAddress)

	// Generate random nonce
	nonce, err := s.generateRandomNonce()
	if err != nil {
		return nil, err
	}

	// Store nonce in Valkey with TTL
	key := s.adminNonceKey(normalizedAddress)
	cmd := s.valkeyClient.B().Set().Key(key).Value(nonce).Ex(s.nonceTTL).Build()
	if err := s.valkeyClient.Do(ctx, cmd).Error(); err != nil {
		return nil, fmt.Errorf("failed to store admin nonce: %w", err)
	}

	return &AdminNonceResult{
		Nonce:       nonce,
		SignMessage: s.buildSignMessage(nonce),
	}, nil
}

// validateAndDeleteNonce validates the nonce and deletes it from Valkey
func (s *AdminAuthService) validateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error {
	key := s.adminNonceKey(walletAddress)

	// Get stored nonce
	getCmd := s.valkeyClient.B().Get().Key(key).Build()
	storedNonce, err := s.valkeyClient.Do(ctx, getCmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return ErrNonceNotFound
		}
		return fmt.Errorf("failed to get admin nonce: %w", err)
	}

	// Validate nonce
	if storedNonce != nonce {
		return ErrNonceMismatch
	}

	// Delete nonce after successful validation (one-time use)
	delCmd := s.valkeyClient.B().Del().Key(key).Build()
	if err := s.valkeyClient.Do(ctx, delCmd).Error(); err != nil {
		return fmt.Errorf("failed to delete admin nonce: %w", err)
	}

	return nil
}

// Login authenticates an admin user with wallet signature
// Returns AdminLoginResult on success, or an error with optional RateLimitInfo
func (s *AdminAuthService) Login(ctx context.Context, walletAddress, signature, nonce string) (*AdminLoginResult, *RateLimitInfo, error) {
	// Normalize wallet address
	normalizedAddress := s.authService.NormalizeWalletAddress(walletAddress)

	// 1. Check rate limit
	allowed, remaining, retryAfter, err := s.rateLimitService.CheckRateLimit(ctx, normalizedAddress)
	if err != nil {
		return nil, nil, err
	}

	if !allowed {
		return nil, &RateLimitInfo{Remaining: 0, RetryAfter: retryAfter}, ErrRateLimitExceeded
	}

	// 2. Validate and delete nonce
	if err := s.validateAndDeleteNonce(ctx, normalizedAddress, nonce); err != nil {
		if errors.Is(err, ErrNonceNotFound) || errors.Is(err, ErrNonceMismatch) {
			return nil, &RateLimitInfo{Remaining: remaining}, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// 3. Verify signature
	signMessage := s.buildSignMessage(nonce)
	if err := s.authService.VerifySignature(normalizedAddress, signature, signMessage); err != nil {
		return nil, &RateLimitInfo{Remaining: remaining}, ErrInvalidCredentials
	}

	// 4. Get admin by wallet address
	admin, err := s.adminRepo.GetByWalletAddress(ctx, normalizedAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &RateLimitInfo{Remaining: remaining}, ErrAdminNotFound
		}
		return nil, nil, err
	}

	// 5. Check if account is active
	if !admin.IsActive {
		return nil, &RateLimitInfo{Remaining: remaining}, ErrAccountInactive
	}

	// 6. Reset rate limit on successful login
	err = s.rateLimitService.ResetRateLimit(ctx, normalizedAddress)
	if err != nil {
		// Log the error but don't fail the login
	}

	// 7. Update last login timestamp
	err = s.adminRepo.UpdateLastLogin(ctx, admin.ID)
	if err != nil {
		// Log the error but don't fail the login
	}

	// 8. Generate JWT token
	token, err := s.adminJwtUtil.GenerateToken(admin.ID, admin.WalletAddress, admin.Role)
	if err != nil {
		return nil, nil, err
	}

	return &AdminLoginResult{
		Token: token,
		Admin: admin,
	}, nil, nil
}
