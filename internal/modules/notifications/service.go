package notifications

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
	nType := strings.TrimSpace(req.Type)
	if nType == "" {
		nType = "info"
	}

	n := &models.Notification{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         strings.TrimSpace(req.Title),
		Description:   strings.TrimSpace(req.Description),
		Type:          nType,
		IsRead:        false,
		ReferenceType: strings.TrimSpace(req.ReferenceType),
		ReferenceID:   strings.TrimSpace(req.ReferenceID),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (
			id, user_id, title, description, type, is_read, read_at,
			reference_type, reference_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		n.ID, n.UserID, n.Title, n.Description, n.Type, n.IsRead, nil,
		n.ReferenceType, n.ReferenceID, n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return toResponse(n), nil
}

func (s *Service) List(ctx context.Context, userID string, isRead *bool) (*ListResponse, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if isRead != nil {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, title, description, type, is_read, read_at,
				reference_type, reference_id, created_at, updated_at
			FROM notifications
			WHERE user_id = $1 AND is_read = $2
			ORDER BY created_at DESC`, userID, *isRead,
		)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, user_id, title, description, type, is_read, read_at,
				reference_type, reference_id, created_at, updated_at
			FROM notifications
			WHERE user_id = $1
			ORDER BY created_at DESC`, userID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items := make([]Response, 0)
	unreadCount := 0
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		if !n.IsRead {
			unreadCount++
		}
		items = append(items, *toResponse(n))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// When filtering, still report total unread for the user.
	if isRead != nil {
		if err := s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID,
		).Scan(&unreadCount); err != nil {
			return nil, fmt.Errorf("count unread: %w", err)
		}
	}

	return &ListResponse{Items: items, UnreadCount: unreadCount}, nil
}

func (s *Service) GetAndMarkRead(ctx context.Context, userID, id string) (*Response, error) {
	n, err := s.findOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if !n.IsRead {
		now := time.Now()
		n.IsRead = true
		n.ReadAt = &now
		n.UpdatedAt = now
		_, err = s.db.ExecContext(ctx, `
			UPDATE notifications
			SET is_read = TRUE, read_at = $3, updated_at = $4
			WHERE id = $1 AND user_id = $2`,
			id, userID, n.ReadAt, n.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("mark notification read: %w", err)
		}
	}

	return toResponse(n), nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) (int, error) {
	now := time.Now()
	res, err := s.db.ExecContext(ctx, `
		UPDATE notifications
		SET is_read = TRUE, read_at = COALESCE(read_at, $2), updated_at = $2
		WHERE user_id = $1 AND is_read = FALSE`,
		userID, now,
	)
	if err != nil {
		return 0, fmt.Errorf("mark all read: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notifications WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
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

func (s *Service) findOwned(ctx context.Context, userID, id string) (*models.Notification, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, description, type, is_read, read_at,
			reference_type, reference_id, created_at, updated_at
		FROM notifications
		WHERE id = $1 AND user_id = $2`, id, userID,
	)
	n, err := scanNotification(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, utils.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find notification: %w", err)
	}
	return n, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanNotification(row scannable) (*models.Notification, error) {
	var n models.Notification
	var readAt sql.NullTime
	err := row.Scan(
		&n.ID, &n.UserID, &n.Title, &n.Description, &n.Type, &n.IsRead, &readAt,
		&n.ReferenceType, &n.ReferenceID, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if readAt.Valid {
		t := readAt.Time
		n.ReadAt = &t
	}
	return &n, nil
}

func toResponse(n *models.Notification) *Response {
	resp := &Response{
		ID:            n.ID,
		Title:         n.Title,
		Description:   n.Description,
		Type:          n.Type,
		IsRead:        n.IsRead,
		ReferenceType: n.ReferenceType,
		ReferenceID:   n.ReferenceID,
		CreatedAt:     n.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     n.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if n.ReadAt != nil {
		s := n.ReadAt.UTC().Format(time.RFC3339)
		resp.ReadAt = &s
	}
	return resp
}
