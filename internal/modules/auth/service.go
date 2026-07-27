package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"softixa-solutions.com/studentify/internal/models"
	"softixa-solutions.com/studentify/internal/utils"
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
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, utils.ErrInvalidCredentials
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
		IsVerified:   u.IsVerified,
		IsActive:     u.IsActive,
		Token:        token,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	expireAt := now.Add(5 * time.Minute)
	otp, err := utils.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}

	u := &models.User{
		ID:           uuid.New().String(),
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		Password:     string(hashed),
		ProfileImage: req.ProfileImage,
		OTP:          otp,
		OTPExpireAt:  expireAt,
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

	if err := utils.SendOTP(u.Email, otp); err != nil {
		return nil, fmt.Errorf("send otp: %w", err)
	}
	successMsg := "OTP sent successfully"
	return &successMsg, nil
}

func (s *Service) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*Response, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if u.OTP == "" || u.OTP != req.OTP {
		return nil, utils.ErrInvalidOTP
	}

	if u.OTPExpireAt.IsZero() || u.OTPExpireAt.Before(time.Now()) {
		return nil, utils.ErrOTPExpired
	}

	if err := s.repo.MarkVerified(ctx, u.ID); err != nil {
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
		IsVerified:   true,
		IsActive:     u.IsActive,
		Token:        token,
	}, nil
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (*string, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}

	expireAt := time.Now().Add(5 * time.Minute)
	if err := s.repo.SaveOTP(ctx, u.ID, otp, expireAt); err != nil {
		return nil, err
	}

	if err := utils.SendOTP(u.Email, otp); err != nil {
		return nil, fmt.Errorf("send otp: %w", err)
	}

	successMsg := "OTP sent successfully"
	return &successMsg, nil
}

func (s *Service) VerifyForgotPasswordOTP(ctx context.Context, req VerifyOTPRequest) (*ForgotPasswordOTPResponse, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if u.OTP == "" || u.OTP != req.OTP {
		return nil, utils.ErrInvalidOTP
	}

	if u.OTPExpireAt.IsZero() || u.OTPExpireAt.Before(time.Now()) {
		return nil, utils.ErrOTPExpired
	}

	resetToken, err := s.tokens.GenerateResetToken(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("generate reset token: %w", err)
	}

	if err := s.repo.ClearOTP(ctx, u.ID); err != nil {
		return nil, err
	}

	return &ForgotPasswordOTPResponse{
		Email:      u.Email,
		ResetToken: resetToken,
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest, resetToken string) (*string, error) {
	claims, err := s.tokens.ParseResetToken(resetToken)
	if err != nil {
		return nil, utils.ErrInvalidResetToken
	}

	if !strings.EqualFold(claims.Email, req.Email) {
		return nil, utils.ErrInvalidResetToken
	}

	u, err := s.repo.FindByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if u.ID != claims.UserID {
		return nil, utils.ErrInvalidResetToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, u.ID, string(hashed)); err != nil {
		return nil, err
	}

	successMsg := "Password updated successfully"
	return &successMsg, nil
}
