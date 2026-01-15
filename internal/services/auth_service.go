package services

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

var (
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrInvalidWalletFormat = errors.New("invalid wallet address format")
	ErrSignatureMismatch   = errors.New("signature does not match wallet address")
)

// AuthServiceInterface defines the interface for authentication operations
type AuthServiceInterface interface {
	VerifySignature(walletAddress, signature, message string) error
	NormalizeWalletAddress(walletAddress string) string
}

type AuthService struct {
	eip712Name    string
	eip712Version string
	eip712ChainID *big.Int
}

func NewAuthService(cfg *config.AuthConfig) *AuthService {
	return &AuthService{
		eip712Name:    cfg.EIP712Name,
		eip712Version: cfg.EIP712Version,
		eip712ChainID: big.NewInt(cfg.EIP712ChainID),
	}
}

func (s *AuthService) VerifySignature(walletAddress, signature, message string) error {
	// Validate wallet address format
	if !common.IsHexAddress(walletAddress) {
		return ErrInvalidWalletFormat
	}

	// Decode signature
	sigBytes := common.FromHex(signature)
	if len(sigBytes) != 65 {
		return ErrInvalidSignature
	}

	// Adjust recovery id (v) for Ethereum signatures
	// Ethereum uses v = 27 or 28, but crypto.Ecrecover expects v = 0 or 1
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Build EIP-712 typed data hash
	typedDataHash, err := s.buildTypedDataHash(message)
	if err != nil {
		return fmt.Errorf("failed to build typed data hash: %w", err)
	}

	// Recover public key from signature
	pubKey, err := crypto.Ecrecover(typedDataHash, sigBytes)
	if err != nil {
		return ErrInvalidSignature
	}

	// Convert public key to address
	pubKeyECDSA, err := crypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return ErrInvalidSignature
	}
	recoveredAddress := crypto.PubkeyToAddress(*pubKeyECDSA)

	// Compare addresses (case-insensitive)
	if !strings.EqualFold(recoveredAddress.Hex(), walletAddress) {
		return ErrSignatureMismatch
	}

	return nil
}

func (s *AuthService) buildTypedDataHash(message string) ([]byte, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
			},
			"Login": []apitypes.Type{
				{Name: "message", Type: "string"},
			},
		},
		PrimaryType: "Login",
		Domain: apitypes.TypedDataDomain{
			Name:    s.eip712Name,
			Version: s.eip712Version,
			ChainId: (*math.HexOrDecimal256)(s.eip712ChainID),
		},
		Message: apitypes.TypedDataMessage{
			"message": message,
		},
	}

	// Hash the domain separator
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain separator: %w", err)
	}

	// Hash the message
	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message: %w", err)
	}

	// Combine: keccak256("\x19\x01" + domainSeparator + messageHash)
	rawData := append([]byte("\x19\x01"), domainSeparator...)
	rawData = append(rawData, messageHash...)

	return crypto.Keccak256(rawData), nil
}

func (s *AuthService) NormalizeWalletAddress(walletAddress string) string {
	return common.HexToAddress(walletAddress).Hex()
}
