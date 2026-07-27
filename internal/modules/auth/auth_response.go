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

type VerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	UserID       string `json:"userId"`
	Phone        string `json:"phone"`
	FullName     string `json:"fullName"`
	Email        string `json:"email"`
	ProfileImage string `json:"profileImage"`
	Role         string `json:"role"`
	IsVerified   bool   `json:"isVerified"`
	IsActive     bool   `json:"isActive"`
	Token        string `json:"token"`
}

type ForgotPasswordOTPResponse struct {
	Email      string `json:"email"`
	ResetToken string `json:"resetToken"`
}
