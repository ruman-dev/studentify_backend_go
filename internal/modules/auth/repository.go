package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"softixa-solutions.com/studentify/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, phone, name, email, profile_image, role, password
		FROM users
		WHERE email = $1`, email,
	).Scan(&u.ID, &u.Phone, &u.FullName, &u.Email, &u.ProfileImage, &u.Role, &u.Password)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return &u, nil
}

func (r *Repository) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (
			id, name, email, phone, password, profile_image,
			is_verified, is_active, is_deleted, is_banned, is_suspended,
			is_locked, is_expired, role, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
		)`,
		u.ID, u.FullName, u.Email, u.Phone, u.Password, u.ProfileImage,
		u.IsVerified, u.IsActive, u.IsDeleted, u.IsBanned, u.IsSuspended,
		u.IsLocked, u.IsExpired, u.Role, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
