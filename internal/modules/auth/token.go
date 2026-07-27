package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const resetTokenPurpose = "password_reset"

type Claims struct {
	UserID string `json:"id"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type ResetClaims struct {
	UserID  string `json:"userId"`
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
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

func (t *TokenManager) GenerateResetToken(userID, email string) (string, error) {
	claims := ResetClaims{
		UserID:  userID,
		Email:   email,
		Purpose: resetTokenPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign reset token: %w", err)
	}
	return signed, nil
}

func (t *TokenManager) ParseResetToken(tokenStr string) (*ResetClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &ResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ResetClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid reset token")
	}
	if claims.Purpose != resetTokenPurpose {
		return nil, fmt.Errorf("invalid reset token purpose")
	}
	return claims, nil
}
