package errors

import "net/http"

var (
	ErrBadRequest            = New("BAD_REQUEST", "invalid request payload", http.StatusBadRequest)
	ErrValidation            = New("VALIDATION_FAILED", "validation failed", http.StatusBadRequest)
	ErrInvalidCredentials    = New("INVALID_CREDENTIALS", "invalid email or password", http.StatusUnauthorized)
	ErrUnauthorized          = New("UNAUTHORIZED", "authorization required", http.StatusUnauthorized)
	ErrForbidden             = New("FORBIDDEN", "forbidden", http.StatusForbidden)
	ErrUserAlreadyExists     = New("USER_ALREADY_EXISTS", "user already exists", http.StatusConflict)
	ErrUserNotFound          = New("USER_NOT_FOUND", "user not found", http.StatusNotFound)
	ErrUserInactive          = New("USER_INACTIVE", "user account is inactive", http.StatusForbidden)
	ErrEmailNotVerified      = New("EMAIL_NOT_VERIFIED", "email is not verified", http.StatusForbidden)
	ErrInvalidOTP            = New("INVALID_OTP", "invalid or expired otp", http.StatusBadRequest)
	ErrInvalidTempToken      = New("INVALID_TEMP_TOKEN", "invalid or expired temporary token", http.StatusUnauthorized)
	ErrSessionNotFound       = New("SESSION_NOT_FOUND", "session not found", http.StatusNotFound)
	ErrSessionRevoked        = New("SESSION_REVOKED", "session is revoked", http.StatusUnauthorized)
	ErrRefreshTokenInvalid   = New("REFRESH_TOKEN_INVALID", "invalid refresh token", http.StatusUnauthorized)
	ErrRefreshTokenExpired   = New("REFRESH_TOKEN_EXPIRED", "refresh token has expired", http.StatusUnauthorized)
	ErrTooManyRequests       = New("TOO_MANY_REQUESTS", "too many requests", http.StatusTooManyRequests)
	ErrEmailVerificationCode = New("EMAIL_VERIFICATION_FAILED", "email verification failed", http.StatusBadRequest)
)
