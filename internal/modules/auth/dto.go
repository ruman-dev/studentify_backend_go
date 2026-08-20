package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	FullName     string `json:"full_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone" validate:"required,phone"`
	Password     string `json:"password" validate:"required,password"`
	ProfileImage string `json:"profile_image"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
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
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type ForgotPasswordOTPResponse struct {
	Email      string `json:"email"`
	ResetToken string `json:"resetToken"`
}
