package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Password     string    `json:"password"`
	ProfileImage string    `json:"profile_image"`
	IsVerified   bool      `json:"is_verified"`
	IsActive     bool      `json:"is_active"`
	IsDeleted    bool      `json:"is_deleted"`
	IsBanned     bool      `json:"is_banned"`
	IsSuspended  bool      `json:"is_suspended"`
	IsLocked     bool      `json:"is_locked"`
	IsExpired    bool      `json:"is_expired"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AuthResponse struct {
	UserID       string `json:"userId"`
	Phone        string `json:"phone"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	ProfileImage string `json:"profileImage"`
	Role         string `json:"role"`
	Token        string `json:"token"`
}
