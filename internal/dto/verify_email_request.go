package dto

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email,max=320"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}
