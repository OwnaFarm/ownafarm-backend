package services

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ownafarm/ownafarm-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_VerifySignature_InvalidWalletFormat(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	// Test with invalid wallet address format
	err := authService.VerifySignature("invalid-wallet", "0x123", "test message")
	assert.ErrorIs(t, err, ErrInvalidWalletFormat)
}

func TestAuthService_VerifySignature_InvalidSignatureLength(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	walletAddress := "0x1234567890abcdef1234567890abcdef12345678"

	// Test with a signature that's too short
	err := authService.VerifySignature(walletAddress, "0x1234", "test message")
	assert.ErrorIs(t, err, ErrInvalidSignature)

	// Test with an empty signature
	err = authService.VerifySignature(walletAddress, "", "test message")
	assert.ErrorIs(t, err, ErrInvalidSignature)
}

func TestAuthService_NormalizeWalletAddress(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase address",
			input:    "0xabcdef1234567890abcdef1234567890abcdef12",
			expected: common.HexToAddress("0xabcdef1234567890abcdef1234567890abcdef12").Hex(),
		},
		{
			name:     "uppercase address",
			input:    "0xABCDEF1234567890ABCDEF1234567890ABCDEF12",
			expected: common.HexToAddress("0xABCDEF1234567890ABCDEF1234567890ABCDEF12").Hex(),
		},
		{
			name:     "mixed case address",
			input:    "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
			expected: common.HexToAddress("0xAbCdEf1234567890AbCdEf1234567890AbCdEf12").Hex(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authService.NormalizeWalletAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthService_BuildTypedDataHash(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	// Test that buildTypedDataHash returns consistent results
	message := "Sign this message to login to OwnaFarm.\n\nNonce: abc123"

	hash1, err := authService.buildTypedDataHash(message)
	assert.NoError(t, err)
	assert.NotNil(t, hash1)
	assert.Len(t, hash1, 32) // Keccak256 produces 32 bytes

	// Same message should produce same hash
	hash2, err := authService.buildTypedDataHash(message)
	assert.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different message should produce different hash
	hash3, err := authService.buildTypedDataHash("different message")
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestAuthService_VerifySignature_Success(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	// Generate a new private key for testing
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Get the wallet address from the private key
	walletAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Build the message
	message := "Sign this message to login to OwnaFarm.\n\nNonce: test-nonce-123"

	// Build typed data hash
	typedDataHash, err := authService.buildTypedDataHash(message)
	assert.NoError(t, err)

	// Sign the hash with the private key
	signature, err := crypto.Sign(typedDataHash, privateKey)
	assert.NoError(t, err)

	// Adjust v value for Ethereum (add 27)
	signature[64] += 27

	// Convert signature to hex string
	signatureHex := "0x" + common.Bytes2Hex(signature)

	// Verify the signature
	err = authService.VerifySignature(walletAddress, signatureHex, message)
	assert.NoError(t, err)
}

func TestAuthService_VerifySignature_WrongWalletAddress(t *testing.T) {
	cfg := &config.AuthConfig{
		EIP712Name:    "OwnaFarm",
		EIP712Version: "1",
		EIP712ChainID: 5003,
	}
	authService := NewAuthService(cfg)

	// Generate a private key for signing
	privateKey, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// Use a different wallet address (not matching the private key)
	wrongWalletAddress := "0x1234567890abcdef1234567890abcdef12345678"

	// Build the message
	message := "Sign this message to login to OwnaFarm.\n\nNonce: test-nonce-456"

	// Build typed data hash and sign
	typedDataHash, err := authService.buildTypedDataHash(message)
	assert.NoError(t, err)

	signature, err := crypto.Sign(typedDataHash, privateKey)
	assert.NoError(t, err)
	signature[64] += 27

	signatureHex := "0x" + common.Bytes2Hex(signature)

	// Verify should fail because wallet address doesn't match
	err = authService.VerifySignature(wrongWalletAddress, signatureHex, message)
	assert.ErrorIs(t, err, ErrSignatureMismatch)
}
