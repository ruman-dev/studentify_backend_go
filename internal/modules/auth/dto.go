package auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	ProfileImage string `json:"profile_image"`
}

type Response struct {
	UserID       string `json:"userId"`
	Phone        string `json:"phone"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	ProfileImage string `json:"profileImage"`
	Role         string `json:"role"`
	Token        string `json:"token"`
}
