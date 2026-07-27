package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"softixa-solutions.com/studentify/internal/models"
	"softixa-solutions.com/studentify/internal/utils"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	var otp sql.NullString
	var otpExpireAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
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

func (r *Repository) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx, `
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
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *Repository) MarkVerified(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
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

func (r *Repository) SaveOTP(ctx context.Context, userID, otp string, expireAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
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

func (r *Repository) ClearOTP(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `
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

func (r *Repository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
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
