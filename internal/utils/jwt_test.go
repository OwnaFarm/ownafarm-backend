package utils

import (
	"testing"
	"time"

	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTUtil_GenerateAndValidateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}
	jwtUtil := NewJWTUtil(cfg)

	userID := "test-user-id-123"
	walletAddress := "0x1234567890abcdef1234567890abcdef12345678"

	// Generate token
	token, err := jwtUtil.GenerateToken(userID, walletAddress)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := jwtUtil.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, walletAddress, claims.WalletAddress)
}

func TestJWTUtil_ValidateToken_InvalidToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}
	jwtUtil := NewJWTUtil(cfg)

	// Test with invalid token
	_, err := jwtUtil.ValidateToken("invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTUtil_ValidateToken_WrongSecret(t *testing.T) {
	cfg1 := &config.JWTConfig{
		Secret:          "secret-key-one-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}
	cfg2 := &config.JWTConfig{
		Secret:          "secret-key-two-for-jwt-testing-32chars",
		ExpirationHours: 24,
	}

	jwtUtil1 := NewJWTUtil(cfg1)
	jwtUtil2 := NewJWTUtil(cfg2)

	// Generate token with first secret
	token, err := jwtUtil1.GenerateToken("user-id", "0x1234")
	require.NoError(t, err)

	// Try to validate with different secret
	_, err = jwtUtil2.ValidateToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestJWTUtil_ValidateToken_ExpiredToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-key-for-jwt-testing-32chars",
		ExpirationHours: 0, // 0 hours = expires immediately
	}
	jwtUtil := NewJWTUtil(cfg)

	// Generate a token that expires immediately
	token, err := jwtUtil.GenerateToken("user-id", "0x1234")
	require.NoError(t, err)

	// Wait a moment to ensure the token is expired
	time.Sleep(100 * time.Millisecond)

	// Validate expired token
	_, err = jwtUtil.ValidateToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}
