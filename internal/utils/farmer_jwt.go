package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

var (
	ErrInvalidFarmerToken = errors.New("invalid farmer token")
	ErrExpiredFarmerToken = errors.New("farmer token has expired")
)

// FarmerClaims represents JWT claims for farmer authentication
type FarmerClaims struct {
	FarmerID      string `json:"farmer_id"`
	WalletAddress string `json:"wallet_address"`
	jwt.RegisteredClaims
}

// FarmerJWTUtil handles JWT operations for farmer authentication
type FarmerJWTUtil struct {
	secret          []byte
	expirationHours int
}

// NewFarmerJWTUtil creates a new FarmerJWTUtil instance
func NewFarmerJWTUtil(cfg *config.JWTConfig) *FarmerJWTUtil {
	return &FarmerJWTUtil{
		secret:          []byte(cfg.Secret),
		expirationHours: cfg.ExpirationHours,
	}
}

// GenerateToken creates a new JWT token for a farmer
func (j *FarmerJWTUtil) GenerateToken(farmerID, walletAddress string) (string, error) {
	now := time.Now()
	claims := FarmerClaims{
		FarmerID:      farmerID,
		WalletAddress: walletAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates a JWT token and returns the farmer claims
func (j *FarmerJWTUtil) ValidateToken(tokenString string) (*FarmerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &FarmerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidFarmerToken
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredFarmerToken
		}
		return nil, ErrInvalidFarmerToken
	}

	claims, ok := token.Claims.(*FarmerClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidFarmerToken
	}

	return claims, nil
}
