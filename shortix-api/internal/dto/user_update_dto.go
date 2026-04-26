package dto

type EmailChangeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type VerifyOTPRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}
