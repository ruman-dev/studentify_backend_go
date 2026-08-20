package auth

import (
	"context"
	"database/sql"
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
	db     *sql.DB
	tokens *utils.TokenManager
}

func NewService(db *sql.DB, tokens *utils.TokenManager) *Service {
	return &Service{db: db, tokens: tokens}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*Response, error) {
	u, err := s.findByEmail(ctx, req.Email)
	if errors.Is(err, utils.ErrUserNotFound) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, utils.ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.tokens.GeneratePair(u.ID, u.Phone, u.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*string, error) {
	existing, err := s.findByEmail(ctx, req.Email)
	if err == nil {
		if existing.IsVerified {
			return nil, utils.ErrEmailTaken
		}
		return s.resendPendingRegistrationOTP(ctx, existing.ID, req)
	}
	if !errors.Is(err, utils.ErrUserNotFound) {
		return nil, err
	}

	if existing, err := s.findByPhone(ctx, req.Phone); err == nil && existing.ID != "" {
		return nil, utils.ErrPhoneTaken
	} else if err != nil && !errors.Is(err, utils.ErrUserNotFound) {
		return nil, err
	}

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

	if err := utils.SendOTP(req.Email, otp); err != nil {
		return nil, fmt.Errorf("%w: %v", utils.ErrOTPSendFailed, err)
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

	if err := s.create(ctx, u); err != nil {
		return nil, err
	}

	successMsg := "OTP sent successfully"
	return &successMsg, nil
}

func (s *Service) resendPendingRegistrationOTP(ctx context.Context, userID string, req RegisterRequest) (*string, error) {
	if existing, err := s.findByPhone(ctx, req.Phone); err == nil && existing.ID != userID {
		return nil, utils.ErrPhoneTaken
	} else if err != nil && !errors.Is(err, utils.ErrUserNotFound) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("generate otp: %w", err)
	}

	if err := utils.SendOTP(req.Email, otp); err != nil {
		return nil, fmt.Errorf("%w: %v", utils.ErrOTPSendFailed, err)
	}

	if err := s.updatePendingRegistration(ctx, userID, req, string(hashed), otp, time.Now().Add(5*time.Minute)); err != nil {
		return nil, err
	}

	successMsg := "OTP sent successfully"
	return &successMsg, nil
}

func (s *Service) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*Response, error) {
	u, err := s.findByEmail(ctx, req.Email)
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

	if err := s.markVerified(ctx, u.ID); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.tokens.GeneratePair(u.ID, u.Phone, u.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) (*string, error) {
	u, err := s.findByEmail(ctx, req.Email)
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
	if err := s.saveOTP(ctx, u.ID, otp, expireAt); err != nil {
		return nil, err
	}

	if err := utils.SendOTP(u.Email, otp); err != nil {
		return nil, fmt.Errorf("%w: %v", utils.ErrOTPSendFailed, err)
	}

	successMsg := "OTP sent successfully"
	return &successMsg, nil
}

func (s *Service) VerifyForgotPasswordOTP(ctx context.Context, req VerifyOTPRequest) (*ForgotPasswordOTPResponse, error) {
	u, err := s.findByEmail(ctx, req.Email)
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

	if err := s.clearOTP(ctx, u.ID); err != nil {
		return nil, err
	}

	return &ForgotPasswordOTPResponse{
		Email:      u.Email,
		ResetToken: resetToken,
	}, nil
}

func (s *Service) RefreshToken(refreshToken string) (*RefreshTokenResponse, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, utils.ErrInvalidToken
	}

	accessToken, nextRefreshToken, err := s.tokens.GeneratePair(claims.UserID, claims.Phone, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token pair: %w", err)
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: nextRefreshToken,
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

	u, err := s.findByEmail(ctx, req.Email)
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

	if err := s.updatePassword(ctx, u.ID, string(hashed)); err != nil {
		return nil, err
	}

	successMsg := "Password updated successfully"
	return &successMsg, nil
}

func (s *Service) findByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	var otp sql.NullString
	var otpExpireAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, phone, name, email, profile_image, role, password,
			otp, otp_expire_at, is_verified
		FROM users
		WHERE email = $1`, email,
	).Scan(
		&u.ID, &u.Phone, &u.FullName, &u.Email, &u.ProfileImage, &u.Role, &u.Password,
		&otp, &otpExpireAt, &u.IsVerified,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if otp.Valid {
		u.OTP = otp.String
	}
	if otpExpireAt.Valid {
		u.OTPExpireAt = otpExpireAt.Time
	}

	return &u, nil
}

func (s *Service) findByPhone(ctx context.Context, phone string) (*models.User, error) {
	var u models.User
	var otp sql.NullString
	var otpExpireAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, phone, name, email, profile_image, role, password,
			otp, otp_expire_at, is_verified
		FROM users
		WHERE phone = $1`, phone,
	).Scan(
		&u.ID, &u.Phone, &u.FullName, &u.Email, &u.ProfileImage, &u.Role, &u.Password,
		&otp, &otpExpireAt, &u.IsVerified,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	if otp.Valid {
		u.OTP = otp.String
	}
	if otpExpireAt.Valid {
		u.OTPExpireAt = otpExpireAt.Time
	}

	return &u, nil
}

func (s *Service) create(ctx context.Context, u *models.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (
			id, name, email, phone, password, profile_image,
			otp, otp_expire_at,
			is_verified, is_active, is_deleted, is_banned, is_suspended,
			is_locked, is_expired, role, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
		)`,
		u.ID, u.FullName, u.Email, u.Phone, u.Password, u.ProfileImage,
		u.OTP, u.OTPExpireAt,
		u.IsVerified, u.IsActive, u.IsDeleted, u.IsBanned, u.IsSuspended,
		u.IsLocked, u.IsExpired, u.Role, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		if utils.IsUniqueViolation(err, "users_email_key") {
			return fmt.Errorf("%w: email already taken", utils.ErrEmailTaken)
		}
		if utils.IsUniqueViolation(err, "users_phone_key") {
			return fmt.Errorf("%w: phone number already taken", utils.ErrPhoneTaken)
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *Service) markVerified(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET is_verified = true,
			otp = NULL,
			otp_expire_at = NULL,
			updated_at = $2
		WHERE id = $1`,
		userID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("mark user verified: %w", err)
	}
	return nil
}

func (s *Service) updatePendingRegistration(ctx context.Context, userID string, req RegisterRequest, passwordHash, otp string, expireAt time.Time) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET name = $2,
			phone = $3,
			password = $4,
			profile_image = $5,
			otp = $6,
			otp_expire_at = $7,
			is_active = true,
			is_deleted = false,
			updated_at = $8
		WHERE id = $1
			AND email = $9
			AND is_verified = false`,
		userID, req.FullName, req.Phone, passwordHash, req.ProfileImage,
		otp, expireAt, time.Now(), req.Email,
	)
	if err != nil {
		if utils.IsUniqueViolation(err, "users_phone_key") {
			return fmt.Errorf("%w: phone number already taken", utils.ErrPhoneTaken)
		}
		return fmt.Errorf("update pending registration: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check pending registration update: %w", err)
	}
	if affected == 0 {
		return utils.ErrEmailTaken
	}
	return nil
}

func (s *Service) saveOTP(ctx context.Context, userID, otp string, expireAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET otp = $2,
			otp_expire_at = $3,
			updated_at = $4
		WHERE id = $1`,
		userID, otp, expireAt, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("save otp: %w", err)
	}
	return nil
}

func (s *Service) clearOTP(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET otp = NULL,
			otp_expire_at = NULL,
			updated_at = $2
		WHERE id = $1`,
		userID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("clear otp: %w", err)
	}
	return nil
}

func (s *Service) updatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users
		SET password = $2,
			otp = NULL,
			otp_expire_at = NULL,
			updated_at = $3
		WHERE id = $1`,
		userID, passwordHash, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
