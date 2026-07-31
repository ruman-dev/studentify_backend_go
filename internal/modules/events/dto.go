package events

type CreateRequest struct {
	SubjectID   *string `json:"subject_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	StartsAt    string  `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
	EventType   string  `json:"event_type"`
}

type UpdateRequest struct {
	SubjectID   *string `json:"subject_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
	StartsAt    *string `json:"starts_at"`
	EndsAt      *string `json:"ends_at"`
	EventType   *string `json:"event_type"`
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
