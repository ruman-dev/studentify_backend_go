package subjects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
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
	teacherID, err := s.normalizeTeacherID(ctx, userID, req.TeacherID)
	if err != nil {
		return nil, err
	}

	scheduleDays := normalizeScheduleDays(req.ScheduleDays)
	startTime := strings.TrimSpace(req.StartTime)
	endTime := strings.TrimSpace(req.EndTime)
	if err := validateScheduleRange(startTime, endTime); err != nil {
		return nil, err
	}

	now := time.Now()
	sub := &models.Subject{
		ID:           uuid.New().String(),
		UserID:       userID,
		TeacherID:    teacherID,
		Name:         strings.TrimSpace(req.Name),
		Code:         strings.TrimSpace(req.Code),
		Description:  strings.TrimSpace(req.Description),
		CreditHours:  req.CreditHours,
		ScheduleDays: scheduleDays,
		StartTime:    startTime,
		EndTime:      endTime,
		Room:         strings.TrimSpace(req.Room),
		MeetingLink:  strings.TrimSpace(req.MeetingLink),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO subjects (
			id, user_id, teacher_id, name, code, description, credit_hours,
			schedule_days, start_time, end_time, room, meeting_link, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		sub.ID, sub.UserID, nullString(sub.TeacherID), sub.Name, sub.Code, sub.Description, nullFloat(sub.CreditHours),
		pq.Array(sub.ScheduleDays), sub.StartTime, sub.EndTime, sub.Room, sub.MeetingLink, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}
	return toResponse(sub), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Response, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, teacher_id, name, code, description, credit_hours,
			schedule_days, start_time, end_time, room, meeting_link, created_at, updated_at
		FROM subjects
		WHERE user_id = $1
		ORDER BY name ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	out := make([]Response, 0)
	for rows.Next() {
		sub, err := scanSubject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *toResponse(sub))
	}
	return out, rows.Err()
}

func (s *Service) Get(ctx context.Context, userID, id string) (*Response, error) {
	sub, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return toResponse(sub), nil
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*Response, error) {
	sub, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if req.TeacherID != nil {
		teacherID, err := s.normalizeTeacherID(ctx, userID, req.TeacherID)
		if err != nil {
			return nil, err
		}
		sub.TeacherID = teacherID
	}
	if req.Name != nil {
		sub.Name = strings.TrimSpace(*req.Name)
	}
	if req.Code != nil {
		sub.Code = strings.TrimSpace(*req.Code)
	}
	if req.Description != nil {
		sub.Description = strings.TrimSpace(*req.Description)
	}
	if req.CreditHours != nil {
		sub.CreditHours = req.CreditHours
	}
	if req.ScheduleDays != nil {
		sub.ScheduleDays = normalizeScheduleDays(req.ScheduleDays)
	}
	if req.StartTime != nil {
		sub.StartTime = strings.TrimSpace(*req.StartTime)
	}
	if req.EndTime != nil {
		sub.EndTime = strings.TrimSpace(*req.EndTime)
	}
	if req.Room != nil {
		sub.Room = strings.TrimSpace(*req.Room)
	}
	if req.MeetingLink != nil {
		sub.MeetingLink = strings.TrimSpace(*req.MeetingLink)
	}
	if err := validateScheduleRange(sub.StartTime, sub.EndTime); err != nil {
		return nil, err
	}
	sub.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE subjects
		SET teacher_id = $3, name = $4, code = $5, description = $6, credit_hours = $7,
			schedule_days = $8, start_time = $9, end_time = $10, room = $11, meeting_link = $12, updated_at = $13
		WHERE id = $1 AND user_id = $2`,
		id, userID, nullString(sub.TeacherID), sub.Name, sub.Code, sub.Description, nullFloat(sub.CreditHours),
		pq.Array(sub.ScheduleDays), sub.StartTime, sub.EndTime, sub.Room, sub.MeetingLink, sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update subject: %w", err)
	}
	return toResponse(sub), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM subjects WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete subject: %w", err)
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

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Subject, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, teacher_id, name, code, description, credit_hours,
			schedule_days, start_time, end_time, room, meeting_link, created_at, updated_at
		FROM subjects
		WHERE id = $1 AND user_id = $2`, id, userID,
	)
	sub, err := scanSubject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find subject: %w", err)
	}
	return sub, nil
}

func (s *Service) normalizeTeacherID(ctx context.Context, userID string, teacherID *string) (*string, error) {
	if teacherID == nil {
		return nil, nil
	}
	id := strings.TrimSpace(*teacherID)
	if id == "" {
		return nil, nil
	}

	var exists string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM teachers WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrInvalidTeacher
	}
	if err != nil {
		return nil, fmt.Errorf("check teacher: %w", err)
	}
	return &id, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSubject(row scannable) (*models.Subject, error) {
	var sub models.Subject
	var teacherID sql.NullString
	var creditHours sql.NullFloat64
	var scheduleDays pq.StringArray
	err := row.Scan(
		&sub.ID, &sub.UserID, &teacherID, &sub.Name, &sub.Code, &sub.Description, &creditHours,
		&scheduleDays, &sub.StartTime, &sub.EndTime, &sub.Room, &sub.MeetingLink, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if teacherID.Valid {
		sub.TeacherID = &teacherID.String
	}
	if creditHours.Valid {
		v := creditHours.Float64
		sub.CreditHours = &v
	}
	sub.ScheduleDays = []string(scheduleDays)
	if sub.ScheduleDays == nil {
		sub.ScheduleDays = []string{}
	}
	return &sub, nil
}

func toResponse(sub *models.Subject) *Response {
	days := sub.ScheduleDays
	if days == nil {
		days = []string{}
	}
	return &Response{
		ID:           sub.ID,
		TeacherID:    sub.TeacherID,
		Name:         sub.Name,
		Code:         sub.Code,
		Description:  sub.Description,
		CreditHours:  sub.CreditHours,
		ScheduleDays: days,
		StartTime:    sub.StartTime,
		EndTime:      sub.EndTime,
		Room:         sub.Room,
		MeetingLink:  sub.MeetingLink,
		CreatedAt:    sub.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    sub.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func normalizeScheduleDays(days []string) []string {
	seen := make(map[string]struct{}, len(days))
	out := make([]string, 0, len(days))
	for _, day := range days {
		d := strings.ToLower(strings.TrimSpace(day))
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		out = append(out, d)
	}
	return out
}

func validateScheduleRange(startTime, endTime string) error {
	if startTime == "" || endTime == "" {
		return nil
	}
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return err
	}
	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return err
	}
	if !end.After(start) {
		return utils.ErrInvalidScheduleTimes
	}
	return nil
}

func nullString(v *string) interface{} {
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
