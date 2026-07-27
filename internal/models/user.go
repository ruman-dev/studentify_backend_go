package models

import "time"

type User struct {
	ID           string
	FullName     string
	Email        string
	Phone        string
	Password     string
	ProfileImage string
	IsVerified   bool
	IsActive     bool
	IsDeleted    bool
	IsBanned     bool
	IsSuspended  bool
	IsLocked     bool
	IsExpired    bool
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
