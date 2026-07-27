package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"softixa-solutions.com/studentify/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already registered")
)

type Service struct {
	repo   *Repository
	tokens *TokenManager
}

func NewService(repo *Repository, tokens *TokenManager) *Service {
	return &Service{repo: repo, tokens: tokens}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*Response, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(u.ID, u.Phone, u.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &Response{
		UserID:       u.ID,
		Phone:        u.Phone,
		FullName:     u.FullName,
		Email:        u.Email,
		ProfileImage: u.ProfileImage,
		Role:         u.Role,
		Token:        token,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*Response, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	u := &models.User{
		ID:           uuid.New().String(),
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     string(hashed),
		ProfileImage: req.ProfileImage,
		IsVerified:   false,
		IsActive:     true,
		IsDeleted:    false,
		IsBanned:     false,
		IsSuspended:  false,
		IsLocked:     false,
		IsExpired:    false,
		Role:         "USER",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.tokens.Generate(u.ID, u.Phone, u.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &Response{
		UserID:       u.ID,
		Phone:        u.Phone,
		FullName:     u.FullName,
		Email:        u.Email,
		ProfileImage: u.ProfileImage,
		Role:         u.Role,
		Token:        token,
	}, nil
}
