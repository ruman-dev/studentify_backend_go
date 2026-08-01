package teachers

type CreateRequest struct {
	FullName    string `json:"full_name" validate:"required"`
	Email       string `json:"email" validate:"softemail"`
	Phone       string `json:"phone" validate:"phone"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	Notes       string `json:"notes"`
}

type UpdateRequest struct {
	FullName    *string `json:"full_name" validate:"omitempty,min=1"`
	Email       *string `json:"email" validate:"omitempty,softemail"`
	Phone       *string `json:"phone" validate:"omitempty,phone"`
	Department  *string `json:"department"`
	Designation *string `json:"designation"`
	Notes       *string `json:"notes"`
}

type Response struct {
	ID          string `json:"id"`
	FullName    string `json:"fullName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	Notes       string `json:"notes"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
