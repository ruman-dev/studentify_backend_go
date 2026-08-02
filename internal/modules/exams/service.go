package exams

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
	if err := s.ensureSubject(ctx, userID, req.SubjectID); err != nil {
		return nil, err
	}

	examDate, err := utils.ParseRFC3339(req.ExamDate)
	if err != nil {
		return nil, fmt.Errorf("parse exam_date: %w", err)
	}

	now := time.Now()
	e := &models.Exam{
		ID:              uuid.New().String(),
		UserID:          userID,
		SubjectID:       strings.TrimSpace(req.SubjectID),
		Title:           strings.TrimSpace(req.Title),
		ExamType:        strings.TrimSpace(req.ExamType),
		ExamDate:        examDate,
		DurationMinutes: req.DurationMinutes,
		Venue:           strings.TrimSpace(req.Venue),
		TotalMarks:      req.TotalMarks,
		ObtainedMarks:   req.ObtainedMarks,
		Notes:           strings.TrimSpace(req.Notes),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO exams (
			id, user_id, subject_id, title, exam_type, exam_date, duration_minutes, venue,
			total_marks, obtained_marks, notes, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.UserID, e.SubjectID, e.Title, e.ExamType, e.ExamDate, nullInt(e.DurationMinutes), e.Venue,
		nullFloat(e.TotalMarks), nullFloat(e.ObtainedMarks), e.Notes, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create exam: %w", err)
	}
	if err := s.ensureCustomExamType(ctx, userID, e.ExamType); err != nil {
		return nil, err
	}
	return toResponse(e), nil
}

func (s *Service) List(ctx context.Context, userID, subjectID string) ([]Response, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if subjectID != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, subject_id, title, exam_type, exam_date, duration_minutes, venue,
				total_marks, obtained_marks, notes, created_at, updated_at
			FROM exams
			WHERE user_id = $1 AND subject_id = $2
			ORDER BY exam_date ASC`, userID, subjectID,
		)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, subject_id, title, exam_type, exam_date, duration_minutes, venue,
				total_marks, obtained_marks, notes, created_at, updated_at
			FROM exams
			WHERE user_id = $1
			ORDER BY exam_date ASC`, userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list exams: %w", err)
	}
	defer rows.Close()

	out := make([]Response, 0)
	for rows.Next() {
		e, err := scanExam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *toResponse(e))
	}
	return out, rows.Err()
}

func (s *Service) ListTypes(ctx context.Context, userID string) (*ExamTypesResponse, error) {
	types := make([]string, 0, len(models.DefaultExamTypes)+8)
	seen := make(map[string]struct{}, len(models.DefaultExamTypes)+8)
	for _, name := range models.DefaultExamTypes {
		types = append(types, name)
		seen[strings.ToLower(name)] = struct{}{}
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT name
		FROM exam_types
		WHERE user_id = $1
		ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list exam types: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		types = append(types, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ExamTypesResponse{Types: types}, nil
}

func (s *Service) ensureCustomExamType(ctx context.Context, userID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	for _, d := range models.DefaultExamTypes {
		if strings.EqualFold(d, name) {
			return nil
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO exam_types (id, user_id, name, created_at)
		SELECT $1, $2, $3, $4
		WHERE NOT EXISTS (
			SELECT 1 FROM exam_types WHERE user_id = $2 AND lower(name) = lower($3)
		)`,
		uuid.New().String(), userID, name, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("save exam type: %w", err)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Response, error) {
	e, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return toResponse(e), nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*Response, error) {
	e, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if req.SubjectID != nil {
		subjectID := strings.TrimSpace(*req.SubjectID)
		if err := s.ensureSubject(ctx, userID, subjectID); err != nil {
			return nil, err
		}
		e.SubjectID = subjectID
	}
	if req.Title != nil {
		e.Title = strings.TrimSpace(*req.Title)
	}
	if req.ExamType != nil {
		e.ExamType = strings.TrimSpace(*req.ExamType)
	}
	if req.ExamDate != nil {
		examDate, err := utils.ParseRFC3339(*req.ExamDate)
		if err != nil {
			return nil, fmt.Errorf("parse exam_date: %w", err)
		}
		e.ExamDate = examDate
	}
	if req.DurationMinutes != nil {
		e.DurationMinutes = req.DurationMinutes
	}
	if req.Venue != nil {
		e.Venue = strings.TrimSpace(*req.Venue)
	}
	if req.TotalMarks != nil {
		e.TotalMarks = req.TotalMarks
	}
	if req.ObtainedMarks != nil {
		e.ObtainedMarks = req.ObtainedMarks
	}
	if req.Notes != nil {
		e.Notes = strings.TrimSpace(*req.Notes)
	}
	e.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE exams
		SET subject_id = $3, title = $4, exam_type = $5, exam_date = $6, duration_minutes = $7, venue = $8,
			total_marks = $9, obtained_marks = $10, notes = $11, updated_at = $12
		WHERE id = $1 AND user_id = $2`,
		id, userID, e.SubjectID, e.Title, e.ExamType, e.ExamDate, nullInt(e.DurationMinutes), e.Venue,
		nullFloat(e.TotalMarks), nullFloat(e.ObtainedMarks), e.Notes, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update exam: %w", err)
	}
	if req.ExamType != nil {
		if err := s.ensureCustomExamType(ctx, userID, e.ExamType); err != nil {
			return nil, err
		}
	}
	return toResponse(e), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM exams WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete exam: %w", err)
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

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Exam, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, subject_id, title, exam_type, exam_date, duration_minutes, venue,
			total_marks, obtained_marks, notes, created_at, updated_at
		FROM exams
		WHERE id = $1 AND user_id = $2`, id, userID,
	)
	e, err := scanExam(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find exam: %w", err)
	}
	return e, nil
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

type scannable interface {
	Scan(dest ...any) error
}

func scanExam(row scannable) (*models.Exam, error) {
	var e models.Exam
	var duration sql.NullInt64
	var total, obtained sql.NullFloat64
	err := row.Scan(
		&e.ID, &e.UserID, &e.SubjectID, &e.Title, &e.ExamType, &e.ExamDate, &duration, &e.Venue,
		&total, &obtained, &e.Notes, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if duration.Valid {
		v := int(duration.Int64)
		e.DurationMinutes = &v
	}
	if total.Valid {
		v := total.Float64
		e.TotalMarks = &v
	}
	if obtained.Valid {
		v := obtained.Float64
		e.ObtainedMarks = &v
	}
	return &e, nil
}

func toResponse(e *models.Exam) *Response {
	return &Response{
		ID:              e.ID,
		SubjectID:       e.SubjectID,
		Title:           e.Title,
		ExamType:        e.ExamType,
		ExamDate:        e.ExamDate.UTC().Format(time.RFC3339),
		DurationMinutes: e.DurationMinutes,
		Venue:           e.Venue,
		TotalMarks:      e.TotalMarks,
		ObtainedMarks:   e.ObtainedMarks,
		Notes:           e.Notes,
		CreatedAt:       e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func nullInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullFloat(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
