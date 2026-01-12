package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ownafarm/ownafarm-backend/internal/config"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID        string `json:"user_id"`
	WalletAddress string `json:"wallet_address"`
	jwt.RegisteredClaims
}

type JWTUtil struct {
	secret          []byte
	expirationHours int
}

func NewJWTUtil(cfg *config.JWTConfig) *JWTUtil {
	return &JWTUtil{
		secret:          []byte(cfg.Secret),
		expirationHours: cfg.ExpirationHours,
	}
}

func (j *JWTUtil) GenerateToken(userID, walletAddress string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:        userID,
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

func (j *JWTUtil) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
