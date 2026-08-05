package attendance

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

func (s *Service) Mark(ctx context.Context, userID string, req MarkRequest) (*RecordResponse, error) {
	subjectID := strings.TrimSpace(req.SubjectID)
	if err := s.ensureOwnedSubject(ctx, userID, subjectID); err != nil {
		return nil, err
	}

	startsAt, err := parseRFC3339(req.SessionStartsAt)
	if err != nil {
		return nil, fmt.Errorf("%w: session_starts_at must be RFC3339", utils.ErrInvalidInput)
	}
	startsAt = truncateToMinute(startsAt.UTC())

	var endsAt *time.Time
	if req.SessionEndsAt != nil && strings.TrimSpace(*req.SessionEndsAt) != "" {
		parsed, err := parseRFC3339(*req.SessionEndsAt)
		if err != nil {
			return nil, fmt.Errorf("%w: session_ends_at must be RFC3339", utils.ErrInvalidInput)
		}
		utc := truncateToMinute(parsed.UTC())
		if !utc.After(startsAt) {
			return nil, fmt.Errorf("%w: session_ends_at must be after session_starts_at", utils.ErrInvalidInput)
		}
		endsAt = &utc
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	note := strings.TrimSpace(req.Note)
	// Notes are for late / absent / excused only.
	if status == "present" {
		note = ""
	}
	now := time.Now().UTC()

	var existingID string
	err = s.db.QueryRowContext(ctx, `
		SELECT id FROM attendance_records
		WHERE user_id = $1 AND subject_id = $2 AND session_starts_at = $3`,
		userID, subjectID, startsAt,
	).Scan(&existingID)

	if err == nil {
		_, err = s.db.ExecContext(ctx, `
			UPDATE attendance_records
			SET status = $3, note = $4, session_ends_at = $5, marked_at = $6, updated_at = $6
			WHERE id = $1 AND user_id = $2`,
			existingID, userID, status, note, nullTime(endsAt), now,
		)
		if err != nil {
			return nil, fmt.Errorf("update attendance: %w", err)
		}
		return s.getRecord(ctx, userID, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check attendance: %w", err)
	}

	id := uuid.New().String()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO attendance_records (
			id, user_id, subject_id, session_starts_at, session_ends_at,
			status, note, marked_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8,$8)`,
		id, userID, subjectID, startsAt, nullTime(endsAt), status, note, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create attendance: %w", err)
	}
	return s.getRecord(ctx, userID, id)
}

func (s *Service) Update(ctx context.Context, userID, id string, req UpdateRequest) (*RecordResponse, error) {
	rec, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		rec.Status = strings.ToLower(strings.TrimSpace(*req.Status))
	}
	if req.Note != nil {
		rec.Note = strings.TrimSpace(*req.Note)
	}
	// Notes are for late / absent / excused only.
	if rec.Status == "present" {
		rec.Note = ""
	}
	rec.UpdatedAt = time.Now().UTC()
	rec.MarkedAt = rec.UpdatedAt

	_, err = s.db.ExecContext(ctx, `
		UPDATE attendance_records
		SET status = $3, note = $4, marked_at = $5, updated_at = $5
		WHERE id = $1 AND user_id = $2`,
		id, userID, rec.Status, rec.Note, rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update attendance: %w", err)
	}
	return s.getRecord(ctx, userID, id)
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM attendance_records WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete attendance: %w", err)
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

func (s *Service) Overview(ctx context.Context, userID string) (*OverviewResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.name, s.code,
			COALESCE(SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN a.status = 'absent' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN a.status = 'late' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN a.status = 'excused' THEN 1 ELSE 0 END), 0)
		FROM subjects s
		LEFT JOIN attendance_records a ON a.subject_id = s.id AND a.user_id = s.user_id
		WHERE s.user_id = $1
		GROUP BY s.id, s.name, s.code
		ORDER BY s.name ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("attendance overview: %w", err)
	}
	defer rows.Close()

	subjects := make([]SubjectSummary, 0)
	totalAttended := 0
	totalClasses := 0

	for rows.Next() {
		var sum SubjectSummary
		if err := rows.Scan(
			&sum.SubjectID, &sum.SubjectName, &sum.SubjectCode,
			&sum.Present, &sum.Absent, &sum.Late, &sum.Excused,
		); err != nil {
			return nil, err
		}
		sum.Attended = sum.Present + sum.Late
		sum.Total = sum.Present + sum.Absent + sum.Late
		sum.Percentage = percentage(sum.Attended, sum.Total)
		subjects = append(subjects, sum)
		totalAttended += sum.Attended
		totalClasses += sum.Total
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &OverviewResponse{
		OverallPercentage: percentage(totalAttended, totalClasses),
		TotalAttended:     totalAttended,
		TotalClasses:      totalClasses,
		Subjects:          subjects,
	}, nil
}

func (s *Service) SubjectDetail(ctx context.Context, userID, subjectID string) (*SubjectDetailResponse, error) {
	if err := s.ensureOwnedSubject(ctx, userID, subjectID); err != nil {
		return nil, err
	}

	overview, err := s.Overview(ctx, userID)
	if err != nil {
		return nil, err
	}

	var summary SubjectSummary
	found := false
	for _, item := range overview.Subjects {
		if item.SubjectID == subjectID {
			summary = item
			found = true
			break
		}
	}
	if !found {
		return nil, utils.ErrNotFound
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.subject_id, s.name, s.code,
			a.session_starts_at, a.session_ends_at, a.status, a.note,
			a.marked_at, a.created_at, a.updated_at
		FROM attendance_records a
		JOIN subjects s ON s.id = a.subject_id
		WHERE a.user_id = $1 AND a.subject_id = $2
		ORDER BY a.session_starts_at DESC, a.marked_at DESC`,
		userID, subjectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list subject attendance: %w", err)
	}
	defer rows.Close()

	records := make([]RecordResponse, 0)
	for rows.Next() {
		resp, err := scanRecordResponse(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, *resp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &SubjectDetailResponse{
		Summary: summary,
		Records: records,
	}, nil
}

func (s *Service) ListInRange(ctx context.Context, userID string, from, to time.Time) ([]RecordResponse, error) {
	if !to.After(from) {
		return nil, fmt.Errorf("%w: to must be after from", utils.ErrInvalidInput)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.id, a.subject_id, s.name, s.code,
			a.session_starts_at, a.session_ends_at, a.status, a.note,
			a.marked_at, a.created_at, a.updated_at
		FROM attendance_records a
		JOIN subjects s ON s.id = a.subject_id
		WHERE a.user_id = $1
			AND a.session_starts_at >= $2
			AND a.session_starts_at < $3
		ORDER BY a.session_starts_at ASC`,
		userID, from.UTC(), to.UTC(),
	)
	if err != nil {
		return nil, fmt.Errorf("list attendance range: %w", err)
	}
	defer rows.Close()

	out := make([]RecordResponse, 0)
	for rows.Next() {
		resp, err := scanRecordResponse(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *resp)
	}
	return out, rows.Err()
}

func (s *Service) getRecord(ctx context.Context, userID, id string) (*RecordResponse, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.subject_id, s.name, s.code,
			a.session_starts_at, a.session_ends_at, a.status, a.note,
			a.marked_at, a.created_at, a.updated_at
		FROM attendance_records a
		JOIN subjects s ON s.id = a.subject_id
		WHERE a.id = $1 AND a.user_id = $2`, id, userID,
	)
	resp, err := scanRecordResponse(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attendance: %w", err)
	}
	return resp, nil
}

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.AttendanceRecord, error) {
	var rec models.AttendanceRecord
	var endsAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, subject_id, session_starts_at, session_ends_at,
			status, note, marked_at, created_at, updated_at
		FROM attendance_records
		WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(
		&rec.ID, &rec.UserID, &rec.SubjectID, &rec.SessionStartsAt, &endsAt,
		&rec.Status, &rec.Note, &rec.MarkedAt, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find attendance: %w", err)
	}
	if endsAt.Valid {
		rec.SessionEndsAt = &endsAt.Time
	}
	return &rec, nil
}

func (s *Service) ensureOwnedSubject(ctx context.Context, userID, subjectID string) error {
	var exists string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM subjects WHERE id = $1 AND user_id = $2`, subjectID, userID,
	).Scan(&exists)
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

func scanRecordResponse(row scannable) (*RecordResponse, error) {
	var (
		id, subjectID, subjectName, subjectCode, status, note string
		startsAt, markedAt, createdAt, updatedAt              time.Time
		endsAt                                                sql.NullTime
	)
	err := row.Scan(
		&id, &subjectID, &subjectName, &subjectCode,
		&startsAt, &endsAt, &status, &note,
		&markedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	resp := &RecordResponse{
		ID:              id,
		SubjectID:       subjectID,
		SubjectName:     subjectName,
		SubjectCode:     subjectCode,
		SessionStartsAt: startsAt.UTC().Format(time.RFC3339),
		Status:          status,
		Note:            note,
		MarkedAt:        markedAt.UTC().Format(time.RFC3339),
		CreatedAt:       createdAt.UTC().Format(time.RFC3339),
		UpdatedAt:       updatedAt.UTC().Format(time.RFC3339),
	}
	if endsAt.Valid {
		v := endsAt.Time.UTC().Format(time.RFC3339)
		resp.SessionEndsAt = &v
	}
	return resp, nil
}

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}

func truncateToMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
}

func nullTime(v *time.Time) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func percentage(attended, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(attended) / float64(total) * 100
}
