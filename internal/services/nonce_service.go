package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/valkey-io/valkey-go"
)

var (
	ErrNonceNotFound = errors.New("nonce not found or expired")
	ErrNonceMismatch = errors.New("nonce mismatch")
)

// NonceServiceInterface defines the interface for nonce operations
type NonceServiceInterface interface {
	GenerateNonce(ctx context.Context, walletAddress string) (string, error)
	ValidateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error
	BuildSignMessage(nonce string) string
}

type NonceService struct {
	client valkey.Client
	ttl    time.Duration
}

func NewNonceService(client valkey.Client, cfg *config.AuthConfig) *NonceService {
	return &NonceService{
		client: client,
		ttl:    time.Duration(cfg.NonceTTLMinutes) * time.Minute,
	}
}

func (s *NonceService) generateRandomNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (s *NonceService) nonceKey(walletAddress string) string {
	return fmt.Sprintf("nonce:%s", strings.ToLower(walletAddress))
}

func (s *NonceService) GenerateNonce(ctx context.Context, walletAddress string) (string, error) {
	nonce, err := s.generateRandomNonce()
	if err != nil {
		return "", err
	}

	key := s.nonceKey(walletAddress)
	cmd := s.client.B().Set().Key(key).Value(nonce).Ex(s.ttl).Build()

	if err := s.client.Do(ctx, cmd).Error(); err != nil {
		return "", fmt.Errorf("failed to store nonce in valkey: %w", err)
	}

	return nonce, nil
}

func (s *NonceService) ValidateAndDeleteNonce(ctx context.Context, walletAddress, nonce string) error {
	key := s.nonceKey(walletAddress)

	// Get the stored nonce
	getCmd := s.client.B().Get().Key(key).Build()
	storedNonce, err := s.client.Do(ctx, getCmd).ToString()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return ErrNonceNotFound
		}
		return fmt.Errorf("failed to get nonce from valkey: %w", err)
	}

	// Validate nonce
	if storedNonce != nonce {
		return ErrNonceMismatch
	}

	// Delete the nonce after successful validation
	delCmd := s.client.B().Del().Key(key).Build()
	if err := s.client.Do(ctx, delCmd).Error(); err != nil {
		return fmt.Errorf("failed to delete nonce from valkey: %w", err)
	}

	return nil
}

func (s *NonceService) BuildSignMessage(nonce string) string {
	return fmt.Sprintf("Sign this message to login to OwnaFarm.\n\nNonce: %s", nonce)
}
