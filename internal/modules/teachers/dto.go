package teachers

type CreateRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Department  string `json:"department"`
	Designation string `json:"designation"`
	Notes       string `json:"notes"`
}

type UpdateRequest struct {
	FullName    *string `json:"full_name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
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
