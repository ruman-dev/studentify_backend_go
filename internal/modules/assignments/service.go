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

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Response, error) {
	if err := s.ensureSubject(ctx, userID, req.SubjectID); err != nil {
		return nil, err
	}

	dueDate, err := utils.ParseRFC3339(req.DueDate)
	if err != nil {
		return nil, fmt.Errorf("parse due_date: %w", err)
	}

	status := normalizeStatus(req.Status)
	total, obtained, err := normalizeMarks(req.TotalMarks, req.ObtainedMarks, status)
	if err != nil {
		return nil, err
	}
	if obtained != nil {
		status = "graded"
	}

	now := time.Now()
	a := &models.Assignment{
		ID:            uuid.New().String(),
		UserID:        userID,
		SubjectID:     strings.TrimSpace(req.SubjectID),
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		DueDate:       dueDate,
		Status:        status,
		TotalMarks:    total,
		ObtainedMarks: obtained,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO assignments (
			id, user_id, subject_id, title, description, due_date, status,
			total_marks, obtained_marks, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.ID, a.UserID, a.SubjectID, a.Title, a.Description, a.DueDate, a.Status,
		nullFloat(a.TotalMarks), nullFloat(a.ObtainedMarks), a.CreatedAt, a.UpdatedAt,
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
	const cols = `id, user_id, subject_id, title, description, due_date, status,
		total_marks, obtained_marks, created_at, updated_at`
	if subjectID != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT `+cols+`
			FROM assignments
			WHERE user_id = $1 AND subject_id = $2
			ORDER BY due_date ASC`, userID, subjectID,
		)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT `+cols+`
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
		a, err := scanAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan assignment: %w", err)
		}
		out = append(out, *toResponse(a))
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

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*Response, error) {
	a, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if req.SubjectID != nil {
		subjectID := strings.TrimSpace(*req.SubjectID)
		if err := s.ensureSubject(ctx, userID, subjectID); err != nil {
			return nil, err
		}
		a.SubjectID = subjectID
	}
	if req.Title != nil {
		a.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		a.Description = strings.TrimSpace(*req.Description)
	}
	if req.DueDate != nil {
		dueDate, err := utils.ParseRFC3339(*req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("parse due_date: %w", err)
		}
		a.DueDate = dueDate
	}
	if req.Status != nil {
		a.Status = normalizeStatus(*req.Status)
	}

	nextTotal := a.TotalMarks
	if req.TotalMarks != nil {
		nextTotal = req.TotalMarks
	}
	nextObtained := a.ObtainedMarks
	if req.ObtainedMarks != nil {
		nextObtained = req.ObtainedMarks
	}
	// Drop score if total marks were removed.
	if nextTotal == nil {
		nextObtained = nil
	}

	total, obtained, err := normalizeMarks(nextTotal, nextObtained, a.Status)
	if err != nil {
		return nil, err
	}
	a.TotalMarks = total
	a.ObtainedMarks = obtained
	if obtained != nil && a.Status != "graded" {
		a.Status = "graded"
	}
	a.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE assignments
		SET subject_id = $3, title = $4, description = $5, due_date = $6, status = $7,
			total_marks = $8, obtained_marks = $9, updated_at = $10
		WHERE id = $1 AND user_id = $2`,
		id, userID, a.SubjectID, a.Title, a.Description, a.DueDate, a.Status,
		nullFloat(a.TotalMarks), nullFloat(a.ObtainedMarks), a.UpdatedAt,
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
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, subject_id, title, description, due_date, status,
			total_marks, obtained_marks, created_at, updated_at
		FROM assignments
		WHERE id = $1 AND user_id = $2`, id, userID,
	)
	a, err := scanAssignment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find assignment: %w", err)
	}
	return a, nil
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

// normalizeMarks validates obtained marks against total marks.
// Obtained marks require a positive total; obtained must be within [0, total].
func normalizeMarks(total, obtained *float64, status string) (*float64, *float64, error) {
	if total != nil {
		if *total <= 0 {
			return nil, nil, fmt.Errorf("%w: total_marks must be greater than 0", utils.ErrInvalidInput)
		}
	}
	if obtained == nil {
		return total, nil, nil
	}
	if total == nil {
		return nil, nil, fmt.Errorf("%w: obtained_marks requires total_marks", utils.ErrInvalidInput)
	}
	if *obtained < 0 || *obtained > *total {
		return nil, nil, fmt.Errorf("%w: obtained_marks must be between 0 and total_marks", utils.ErrInvalidInput)
	}
	_ = status
	return total, obtained, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAssignment(row scannable) (*models.Assignment, error) {
	var a models.Assignment
	var total, obtained sql.NullFloat64
	err := row.Scan(
		&a.ID, &a.UserID, &a.SubjectID, &a.Title, &a.Description, &a.DueDate, &a.Status,
		&total, &obtained, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if total.Valid {
		v := total.Float64
		a.TotalMarks = &v
	}
	if obtained.Valid {
		v := obtained.Float64
		a.ObtainedMarks = &v
	}
	return &a, nil
}

func toResponse(a *models.Assignment) *Response {
	return &Response{
		ID:            a.ID,
		SubjectID:     a.SubjectID,
		Title:         a.Title,
		Description:   a.Description,
		DueDate:       a.DueDate.UTC().Format(time.RFC3339),
		Status:        a.Status,
		TotalMarks:    a.TotalMarks,
		ObtainedMarks: a.ObtainedMarks,
		CreatedAt:     a.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     a.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func nullFloat(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
