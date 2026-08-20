package utils

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidCredentials    = errors.New("Invalid credentials")
	ErrInvalidOTP            = errors.New("Invalid OTP")
	ErrOTPExpired            = errors.New("OTP expired")
	ErrOTPVerificationFailed = errors.New("OTP verification failed")
	ErrUserNotFound          = errors.New("User not found")
	ErrEmailTaken            = errors.New("Email already registered")
	ErrPhoneTaken            = errors.New("Phone already registered")
	ErrInvalidEmail          = errors.New("Invalid email")
	ErrInvalidPhone          = errors.New("Invalid phone")
	ErrInvalidPassword       = errors.New("Invalid password")
	ErrInvalidFullName       = errors.New("Invalid full name")
	ErrInvalidProfileImage   = errors.New("Invalid profile image")
	ErrInvalidRole           = errors.New("Invalid role")
	ErrInvalidToken          = errors.New("Invalid token")
	ErrInvalidResetToken     = errors.New("Invalid reset token")
	ErrOTPSendFailed         = errors.New("Failed to send OTP")
	ErrPasswordMismatch      = errors.New("Passwords do not match")
	ErrUserNotVerified       = errors.New("User not verified")
	ErrNotFound              = errors.New("Resource not found")
	ErrForbidden             = errors.New("Forbidden")
	ErrConflict              = errors.New("Conflict")
	ErrInvalidSubject        = errors.New("Invalid subject")
	ErrInvalidScheduleTimes  = errors.New("end_time must be after start_time")
	ErrInvalidTeacher        = errors.New("Invalid teacher")
	ErrInvalidInput          = errors.New("Invalid input")
)

func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}
