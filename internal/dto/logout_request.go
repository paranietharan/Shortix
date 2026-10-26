package dto

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"omitempty,min=32"`
}
