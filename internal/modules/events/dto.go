package events

type CreateRequest struct {
	SubjectID   *string `json:"subject_id"`
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	StartsAt    string  `json:"starts_at" validate:"required,datetime"`
	EndsAt      *string `json:"ends_at" validate:"omitempty,datetime"`
	EventType   string  `json:"event_type" validate:"omitempty,oneof=class meeting deadline other"`
}

type UpdateRequest struct {
	SubjectID   *string `json:"subject_id"`
	Title       *string `json:"title" validate:"omitempty,min=1"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
	StartsAt    *string `json:"starts_at" validate:"omitempty,datetime"`
	EndsAt      *string `json:"ends_at" validate:"omitempty,datetime"`
	EventType   *string `json:"event_type" validate:"omitempty,oneof=class meeting deadline other"`
}

type Response struct {
	ID          string  `json:"id"`
	SubjectID   *string `json:"subjectId,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	StartsAt    string  `json:"startsAt"`
	EndsAt      *string `json:"endsAt,omitempty"`
	EventType   string  `json:"eventType"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}
