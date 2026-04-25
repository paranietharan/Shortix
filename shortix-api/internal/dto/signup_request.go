package dto

type SignupRequest struct {
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}
