package dto

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email,max=320"`
	TempToken   string `json:"temp_token" binding:"required,min=32"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}
