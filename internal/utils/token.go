package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const resetTokenPurpose = "password_reset"

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type Claims struct {
	UserID    string `json:"id"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

type ResetClaims struct {
	UserID  string `json:"userId"`
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewTokenManager(secret []byte, expiry time.Duration) *TokenManager {
	return &TokenManager{
		secret:        secret,
		accessExpiry:  expiry,
		refreshExpiry: 30 * 24 * time.Hour,
	}
}

func (t *TokenManager) Generate(userID, phone, role string) (string, error) {
	return t.GenerateAccessToken(userID, phone, role)
}

func (t *TokenManager) GeneratePair(userID, phone, role string) (accessToken string, refreshToken string, err error) {
	accessToken, err = t.GenerateAccessToken(userID, phone, role)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = t.GenerateRefreshToken(userID, phone, role)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (t *TokenManager) GenerateAccessToken(userID, phone, role string) (string, error) {
	return t.generateAuthToken(userID, phone, role, accessTokenType, t.accessExpiry)
}

func (t *TokenManager) GenerateRefreshToken(userID, phone, role string) (string, error) {
	return t.generateAuthToken(userID, phone, role, refreshTokenType, t.refreshExpiry)
}

func (t *TokenManager) generateAuthToken(userID, phone, role, tokenType string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		Phone:     phone,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (t *TokenManager) Parse(tokenStr string) (*Claims, error) {
	return t.ParseAccessToken(tokenStr)
}

func (t *TokenManager) ParseAccessToken(tokenStr string) (*Claims, error) {
	claims, err := t.parseAuthToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != "" && claims.TokenType != accessTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (t *TokenManager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	claims, err := t.parseAuthToken(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != refreshTokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (t *TokenManager) parseAuthToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UserID == "" || claims.Phone == "" || claims.Role == "" {
		return nil, ErrInvalidToken
	}
	if claims.TokenType == "" && claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) > t.accessExpiry {
		return nil, ErrInvalidToken
	}
	return claims, nil
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
		return nil, ErrInvalidResetToken
	}
	if claims.Purpose != resetTokenPurpose {
		return nil, ErrInvalidResetToken
	}
	if claims.UserID == "" || claims.Email == "" {
		return nil, ErrInvalidResetToken
	}
	return claims, nil
}
