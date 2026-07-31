package utils

import "errors"

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
	ErrPasswordMismatch      = errors.New("Passwords do not match")
	ErrUserNotVerified       = errors.New("User not verified")
	ErrNotFound              = errors.New("Resource not found")
	ErrForbidden             = errors.New("Forbidden")
	ErrConflict              = errors.New("Conflict")
	ErrInvalidSubject        = errors.New("Invalid subject")
	ErrInvalidTeacher        = errors.New("Invalid teacher")
)
