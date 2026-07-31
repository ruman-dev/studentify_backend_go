package assignments

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

var allowedStatuses = map[string]struct{}{
	"pending":   {},
	"submitted": {},
	"graded":    {},
	"overdue":   {},
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest, dueDate time.Time) (*Response, error) {
	if err := s.ensureSubject(ctx, userID, req.SubjectID); err != nil {
		return nil, err
	}

	status := normalizeStatus(req.Status)
	now := time.Now()
	a := &models.Assignment{
		ID:          uuid.New().String(),
		UserID:      userID,
		SubjectID:   strings.TrimSpace(req.SubjectID),
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		DueDate:     dueDate,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assignments (
			id, user_id, subject_id, title, description, due_date, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.UserID, a.SubjectID, a.Title, a.Description, a.DueDate, a.Status, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create assignment: %w", err)
	}
	return toResponse(a), nil
}

func (s *Service) List(ctx context.Context, userID, subjectID string) ([]Response, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if subjectID != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, subject_id, title, description, due_date, status, created_at, updated_at
			FROM assignments
			WHERE user_id = $1 AND subject_id = $2
			ORDER BY due_date ASC`, userID, subjectID,
		)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, subject_id, title, description, due_date, status, created_at, updated_at
			FROM assignments
			WHERE user_id = $1
			ORDER BY due_date ASC`, userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list assignments: %w", err)
	}
	defer rows.Close()

	out := make([]Response, 0)
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.SubjectID, &a.Title, &a.Description, &a.DueDate, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan assignment: %w", err)
		}
		out = append(out, *toResponse(&a))
	}
	return out, rows.Err()
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Response, error) {
	a, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return toResponse(a), nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest, dueDate time.Time) (*Response, error) {
	a, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureSubject(ctx, userID, req.SubjectID); err != nil {
		return nil, err
	}

	a.SubjectID = strings.TrimSpace(req.SubjectID)
	a.Title = strings.TrimSpace(req.Title)
	a.Description = strings.TrimSpace(req.Description)
	a.DueDate = dueDate
	a.Status = normalizeStatus(req.Status)
	a.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE assignments
		SET subject_id = $3, title = $4, description = $5, due_date = $6, status = $7, updated_at = $8
		WHERE id = $1 AND user_id = $2`,
		id, userID, a.SubjectID, a.Title, a.Description, a.DueDate, a.Status, a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update assignment: %w", err)
	}
	return toResponse(a), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM assignments WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete assignment: %w", err)
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

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Assignment, error) {
	var a models.Assignment
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, subject_id, title, description, due_date, status, created_at, updated_at
		FROM assignments
		WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&a.ID, &a.UserID, &a.SubjectID, &a.Title, &a.Description, &a.DueDate, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find assignment: %w", err)
	}
	return &a, nil
}

func (s *Service) ensureSubject(ctx context.Context, userID, subjectID string) error {
	subjectID = strings.TrimSpace(subjectID)
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM subjects WHERE id = $1 AND user_id = $2`, subjectID, userID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return utils.ErrInvalidSubject
	}
	if err != nil {
		return fmt.Errorf("check subject: %w", err)
	}
	return nil
}

func normalizeStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "pending"
	}
	if _, ok := allowedStatuses[status]; ok {
		return status
	}
	return "pending"
}

func IsValidStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return true
	}
	_, ok := allowedStatuses[status]
	return ok
}

func toResponse(a *models.Assignment) *Response {
	return &Response{
		ID:          a.ID,
		SubjectID:   a.SubjectID,
		Title:       a.Title,
		Description: a.Description,
		DueDate:     a.DueDate.UTC().Format(time.RFC3339),
		Status:      a.Status,
		CreatedAt:   a.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
