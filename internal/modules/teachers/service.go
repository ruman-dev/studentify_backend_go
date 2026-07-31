package teachers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"softixa-solutions.com/studentify/internal/models"
	"softixa-solutions.com/studentify/internal/utils"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Response, error) {
	now := time.Now()
	t := &models.Teacher{
		ID:          uuid.New().String(),
		UserID:      userID,
		FullName:    strings.TrimSpace(req.FullName),
		Email:       strings.TrimSpace(req.Email),
		Phone:       strings.TrimSpace(req.Phone),
		Department:  strings.TrimSpace(req.Department),
		Designation: strings.TrimSpace(req.Designation),
		Notes:       strings.TrimSpace(req.Notes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO teachers (
			id, user_id, full_name, email, phone, department, designation, notes, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.UserID, t.FullName, t.Email, t.Phone, t.Department, t.Designation, t.Notes, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create teacher: %w", err)
	}
	return toResponse(t), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Response, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, full_name, email, phone, department, designation, notes, created_at, updated_at
		FROM teachers
		WHERE user_id = $1
		ORDER BY full_name ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}
	defer rows.Close()

	out := make([]Response, 0)
	for rows.Next() {
		var t models.Teacher
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.FullName, &t.Email, &t.Phone, &t.Department, &t.Designation, &t.Notes, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan teacher: %w", err)
		}
		out = append(out, *toResponse(&t))
	}
	return out, rows.Err()
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Response, error) {
	t, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return toResponse(t), nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*Response, error) {
	t, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if req.FullName != nil {
		t.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Email != nil {
		t.Email = strings.TrimSpace(*req.Email)
	}
	if req.Phone != nil {
		t.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Department != nil {
		t.Department = strings.TrimSpace(*req.Department)
	}
	if req.Designation != nil {
		t.Designation = strings.TrimSpace(*req.Designation)
	}
	if req.Notes != nil {
		t.Notes = strings.TrimSpace(*req.Notes)
	}
	t.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE teachers
		SET full_name = $3, email = $4, phone = $5, department = $6, designation = $7, notes = $8, updated_at = $9
		WHERE id = $1 AND user_id = $2`,
		id, userID, t.FullName, t.Email, t.Phone, t.Department, t.Designation, t.Notes, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update teacher: %w", err)
	}
	return toResponse(t), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM teachers WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete teacher: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Teacher, error) {
	var t models.Teacher
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, full_name, email, phone, department, designation, notes, created_at, updated_at
		FROM teachers
		WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&t.ID, &t.UserID, &t.FullName, &t.Email, &t.Phone, &t.Department, &t.Designation, &t.Notes, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find teacher: %w", err)
	}
	return &t, nil
}

func toResponse(t *models.Teacher) *Response {
	return &Response{
		ID:          t.ID,
		FullName:    t.FullName,
		Email:       t.Email,
		Phone:       t.Phone,
		Department:  t.Department,
		Designation: t.Designation,
		Notes:       t.Notes,
		CreatedAt:   t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
