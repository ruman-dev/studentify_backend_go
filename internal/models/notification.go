package models

import "time"

type Notification struct {
	ID            string
	UserID        string
	Title         string
	Description   string
	Type          string // info, warning, alert, reminder
	IsRead        bool
	ReadAt        *time.Time
	ReferenceType string // e.g. exam, assignment, event
	ReferenceID   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
