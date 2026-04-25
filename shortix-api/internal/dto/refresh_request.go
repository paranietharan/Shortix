package dto

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=32"`
}
