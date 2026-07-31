package subjects

type CreateRequest struct {
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Description string   `json:"description"`
	TeacherID   *string  `json:"teacher_id"`
	CreditHours *float64 `json:"credit_hours"`
}

type UpdateRequest struct {
	Name        *string  `json:"name"`
	Code        *string  `json:"code"`
	Description *string  `json:"description"`
	TeacherID   *string  `json:"teacher_id"`
	CreditHours *float64 `json:"credit_hours"`
}

type Response struct {
	ID          string   `json:"id"`
	TeacherID   *string  `json:"teacherId,omitempty"`
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Description string   `json:"description"`
	CreditHours *float64 `json:"creditHours,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}
