package assignments

type CreateRequest struct {
	SubjectID     string   `json:"subject_id" validate:"required"`
	Title         string   `json:"title" validate:"required"`
	Description   string   `json:"description"`
	DueDate       string   `json:"due_date" validate:"required,datetime"`
	Status        string   `json:"status" validate:"omitempty,oneof=pending submitted graded overdue"`
	TotalMarks    *float64 `json:"total_marks"`
	ObtainedMarks *float64 `json:"obtained_marks"`
}

type UpdateRequest struct {
	SubjectID     *string  `json:"subject_id" validate:"omitempty,min=1"`
	Title         *string  `json:"title" validate:"omitempty,min=1"`
	Description   *string  `json:"description"`
	DueDate       *string  `json:"due_date" validate:"omitempty,datetime"`
	Status        *string  `json:"status" validate:"omitempty,oneof=pending submitted graded overdue"`
	TotalMarks    *float64 `json:"total_marks"`
	ObtainedMarks *float64 `json:"obtained_marks"`
}

type Response struct {
	ID            string   `json:"id"`
	SubjectID     string   `json:"subjectId"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	DueDate       string   `json:"dueDate"`
	Status        string   `json:"status"`
	TotalMarks    *float64 `json:"totalMarks"`
	ObtainedMarks *float64 `json:"obtainedMarks"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}
