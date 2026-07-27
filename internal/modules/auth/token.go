package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"id"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	expiry time.Duration
}

func NewTokenManager(secret []byte, expiry time.Duration) *TokenManager {
	return &TokenManager{secret: secret, expiry: expiry}
}

func (t *TokenManager) Generate(userID, phone, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
