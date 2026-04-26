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

type UpdateProfileRequest struct {
	FirstName         *string `json:"first_name" binding:"omitempty,max=50"`
	LastName          *string `json:"last_name" binding:"omitempty,max=50"`
	ProfilePictureURL *string `json:"profile_picture_url" binding:"omitempty,url"`
	Bio               *string `json:"bio" binding:"omitempty,max=500"`
	PhoneNumber       *string `json:"phone_number" binding:"omitempty,max=20"`
}
