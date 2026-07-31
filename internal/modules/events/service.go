package events

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

var allowedEventTypes = map[string]struct{}{
	"class":    {},
	"meeting":  {},
	"deadline": {},
	"other":    {},
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest, startsAt time.Time, endsAt *time.Time) (*Response, error) {
	subjectID, err := s.normalizeSubjectID(ctx, userID, req.SubjectID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	e := &models.Event{
		ID:          uuid.New().String(),
		UserID:      userID,
		SubjectID:   subjectID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Location:    strings.TrimSpace(req.Location),
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		EventType:   normalizeEventType(req.EventType),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO events (
			id, user_id, subject_id, title, description, location, starts_at, ends_at, event_type, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.UserID, nullString(e.SubjectID), e.Title, e.Description, e.Location, e.StartsAt, nullTime(e.EndsAt), e.EventType, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return toResponse(e), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Response, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, subject_id, title, description, location, starts_at, ends_at, event_type, created_at, updated_at
		FROM events
		WHERE user_id = $1
		ORDER BY starts_at ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	out := make([]Response, 0)
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *toResponse(e))
	}
	return out, rows.Err()
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
		subjectID, err := s.normalizeSubjectID(ctx, userID, req.SubjectID)
		if err != nil {
			return nil, err
		}
		e.SubjectID = subjectID
	}
	if req.Title != nil {
		e.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		e.Description = strings.TrimSpace(*req.Description)
	}
	if req.Location != nil {
		e.Location = strings.TrimSpace(*req.Location)
	}
	if req.StartsAt != nil {
		startsAt, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.StartsAt))
		if err != nil {
			return nil, fmt.Errorf("parse starts_at: %w", err)
		}
		e.StartsAt = startsAt
	}
	if req.EndsAt != nil {
		raw := strings.TrimSpace(*req.EndsAt)
		if raw == "" {
			e.EndsAt = nil
		} else {
			endsAt, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return nil, fmt.Errorf("parse ends_at: %w", err)
			}
			if endsAt.Before(e.StartsAt) {
				return nil, fmt.Errorf("ends_at must be after starts_at")
			}
			e.EndsAt = &endsAt
		}
	}
	if req.EventType != nil {
		e.EventType = normalizeEventType(*req.EventType)
	}
	e.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE events
		SET subject_id = $3, title = $4, description = $5, location = $6,
			starts_at = $7, ends_at = $8, event_type = $9, updated_at = $10
		WHERE id = $1 AND user_id = $2`,
		id, userID, nullString(e.SubjectID), e.Title, e.Description, e.Location,
		e.StartsAt, nullTime(e.EndsAt), e.EventType, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	return toResponse(e), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
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

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Event, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, subject_id, title, description, location, starts_at, ends_at, event_type, created_at, updated_at
		FROM events
		WHERE id = $1 AND user_id = $2`, id, userID,
	)
	e, err := scanEvent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event: %w", err)
	}
	return e, nil
}

func (s *Service) normalizeSubjectID(ctx context.Context, userID string, subjectID *string) (*string, error) {
	if subjectID == nil {
		return nil, nil
	}
	id := strings.TrimSpace(*subjectID)
	if id == "" {
		return nil, nil
	}

	var exists string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM subjects WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrInvalidSubject
	}
	if err != nil {
		return nil, fmt.Errorf("check subject: %w", err)
	}
	return &id, nil
}

func normalizeEventType(eventType string) string {
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	if eventType == "" {
		return "other"
	}
	if _, ok := allowedEventTypes[eventType]; ok {
		return eventType
	}
	return "other"
}

func IsValidEventType(eventType string) bool {
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	if eventType == "" {
		return true
	}
	_, ok := allowedEventTypes[eventType]
	return ok
}

type scannable interface {
	Scan(dest ...any) error
}

func scanEvent(row scannable) (*models.Event, error) {
	var e models.Event
	var subjectID sql.NullString
	var endsAt sql.NullTime
	err := row.Scan(
		&e.ID, &e.UserID, &subjectID, &e.Title, &e.Description, &e.Location, &e.StartsAt, &endsAt, &e.EventType, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if subjectID.Valid {
		e.SubjectID = &subjectID.String
	}
	if endsAt.Valid {
		t := endsAt.Time
		e.EndsAt = &t
	}
	return &e, nil
}

func toResponse(e *models.Event) *Response {
	resp := &Response{
		ID:          e.ID,
		SubjectID:   e.SubjectID,
		Title:       e.Title,
		Description: e.Description,
		Location:    e.Location,
		StartsAt:    e.StartsAt.UTC().Format(time.RFC3339),
		EventType:   e.EventType,
		CreatedAt:   e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   e.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if e.EndsAt != nil {
		s := e.EndsAt.UTC().Format(time.RFC3339)
		resp.EndsAt = &s
	}
	return resp
}

func nullString(v *string) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullTime(v *time.Time) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
