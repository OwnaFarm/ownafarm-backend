package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/valkey-io/valkey-go"
)

// FarmerNonceServiceInterface defines the interface for farmer nonce operations
type FarmerNonceServiceInterface interface {
	GenerateNonce(ctx context.Context, walletAddress string) (string, error)
	ValidateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error
	BuildSignMessage(nonce string) string
}

// FarmerNonceService handles nonce generation and validation for farmer authentication
type FarmerNonceService struct {
	client valkey.Client
	ttl    time.Duration
}

// NewFarmerNonceService creates a new FarmerNonceService instance
func NewFarmerNonceService(client valkey.Client, cfg *config.AuthConfig) *FarmerNonceService {
	return &FarmerNonceService{
		client: client,
		ttl:    time.Duration(cfg.NonceTTLMinutes) * time.Minute,
	}
}

func (s *FarmerNonceService) generateRandomNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *FarmerNonceService) nonceKey(walletAddress string) string {
	// Use separate prefix to avoid collision with investor nonces
	return fmt.Sprintf("farmer_nonce:%s", strings.ToLower(walletAddress))
}

// GenerateNonce generates a new nonce for a farmer wallet address
func (s *FarmerNonceService) GenerateNonce(ctx context.Context, walletAddress string) (string, error) {
	nonce, err := s.generateRandomNonce()
	if err != nil {
		return "", err
	}

	key := s.nonceKey(walletAddress)
	cmd := s.client.B().Set().Key(key).Value(nonce).Ex(s.ttl).Build()

	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return "", fmt.Errorf("failed to store farmer nonce in valkey: %w", err)
	}

	return nonce, nil
}

// ValidateAndDeleteNonce validates a nonce and deletes it after successful validation
func (s *FarmerNonceService) ValidateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error {
	key := s.nonceKey(walletAddress)

	// Get the stored nonce
	getCmd := s.client.B().Get().Key(key).Build()
	storedNonce, err := s.client.Do(ctx, getCmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return ErrNonceNotFound
		}
		return fmt.Errorf("failed to get farmer nonce from valkey: %w", err)
	}

	// Validate nonce
	if storedNonce != nonce {
		return ErrNonceMismatch
	}

	// Delete the nonce after successful validation
	delCmd := s.client.B().Del().Key(key).Build()
	if err := s.client.Do(ctx, delCmd).Error(); err != nil {
		return fmt.Errorf("failed to delete farmer nonce from valkey: %w", err)
	}

	return nil
}

// BuildSignMessage builds the message to be signed by the farmer's wallet
func (s *FarmerNonceService) BuildSignMessage(nonce string) string {
	return fmt.Sprintf("Sign this message to login to OwnaFarm as Farmer.\n\nNonce: %s", nonce)
}
