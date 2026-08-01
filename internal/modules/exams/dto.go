package exams

type CreateRequest struct {
	SubjectID       string   `json:"subject_id" validate:"required"`
	Title           string   `json:"title" validate:"required"`
	ExamDate        string   `json:"exam_date" validate:"required,datetime"`
	DurationMinutes *int     `json:"duration_minutes"`
	Venue           string   `json:"venue"`
	TotalMarks      *float64 `json:"total_marks"`
	ObtainedMarks   *float64 `json:"obtained_marks"`
	Notes           string   `json:"notes"`
}

type UpdateRequest struct {
	SubjectID       *string  `json:"subject_id" validate:"omitempty,min=1"`
	Title           *string  `json:"title" validate:"omitempty,min=1"`
	ExamDate        *string  `json:"exam_date" validate:"omitempty,datetime"`
	DurationMinutes *int     `json:"duration_minutes"`
	Venue           *string  `json:"venue"`
	TotalMarks      *float64 `json:"total_marks"`
	ObtainedMarks   *float64 `json:"obtained_marks"`
	Notes           *string  `json:"notes"`
}

type Response struct {
	ID              string   `json:"id"`
	SubjectID       string   `json:"subjectId"`
	Title           string   `json:"title"`
	ExamDate        string   `json:"examDate"`
	DurationMinutes *int     `json:"durationMinutes,omitempty"`
	Venue           string   `json:"venue"`
	TotalMarks      *float64 `json:"totalMarks,omitempty"`
	ObtainedMarks   *float64 `json:"obtainedMarks,omitempty"`
	Notes           string   `json:"notes"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}
