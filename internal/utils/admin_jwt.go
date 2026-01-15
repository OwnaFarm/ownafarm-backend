package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

// AdminClaims represents JWT claims for admin users
type AdminClaims struct {
	AdminID       string `json:"admin_id"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
	jwt.RegisteredClaims
}

// AdminJWTUtil handles JWT operations for admin users
type AdminJWTUtil struct {
	secret          []byte
	expirationHours int
}

// NewAdminJWTUtil creates a new AdminJWTUtil instance
func NewAdminJWTUtil(cfg *config.JWTConfig) *AdminJWTUtil {
	return &AdminJWTUtil{
		secret:          []byte(cfg.Secret),
		expirationHours: cfg.ExpirationHours,
	}
}

// GenerateToken generates a JWT token for an admin user
func (j *AdminJWTUtil) GenerateToken(adminID, walletAddress, role string) (string, error) {
	now := time.Now()
	claims := AdminClaims{
		AdminID:       adminID,
		WalletAddress: walletAddress,
		Role:          role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(j.expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "ownafarm-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates and parses an admin JWT token
func (j *AdminJWTUtil) ValidateToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify issuer for admin tokens
	if claims.Issuer != "ownafarm-admin" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
