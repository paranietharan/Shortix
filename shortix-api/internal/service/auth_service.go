package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"shortix-api/internal/config"
	"shortix-api/internal/dto"
	apperrors "shortix-api/internal/errors"
	"shortix-api/internal/model"
	"shortix-api/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type RequestMeta struct {
	IP        string
	UserAgent string
	Device    string
}

type AuthService struct {
	users        repository.UserRepository
	sessions     repository.SessionRepository
	otpStore     repository.OTPRepository
	emails       EmailSender
	tokenManager *TokenManager
	cfg          *config.Config
	logger       *slog.Logger
	now          func() time.Time
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	otpStore repository.OTPRepository,
	emails EmailSender,
	tokenManager *TokenManager,
	cfg *config.Config,
	logger *slog.Logger,
) *AuthService {
	if logger == nil {
		logger = slog.Default()
	}
	if emails == nil {
		emails = NewNoopEmailSender(logger)
	}

	return &AuthService{
		users:        users,
		sessions:     sessions,
		otpStore:     otpStore,
		emails:       emails,
		tokenManager: tokenManager,
		cfg:          cfg,
		logger:       logger,
		now:          time.Now,
	}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (*dto.MessageResponse, error) {
	email := normalizeEmail(req.Email)

	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return nil, apperrors.ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.InternalServerError()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		s.logger.Error("hash password failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	user := &model.User{
		Email:           email,
		PasswordHash:    string(hashedPassword),
		Role:            "USER",
		IsActive:        true,
		IsEmailVerified: false,
	}

	if err := s.users.Create(ctx, user); err != nil {
		s.logger.Error("create user failed", "error", err, "email", email)
		return nil, apperrors.InternalServerError()
	}

	otp, err := generateOTP(6)
	if err != nil {
		s.logger.Error("generate signup otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if err := s.otpStore.SetEmailVerificationOTP(ctx, email, otp, s.cfg.EmailVerifyOTPTTL); err != nil {
		s.logger.Error("store signup otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	s.sendEmailBestEffort(ctx, email, "Verify your Shortix email", "email_verification_otp.html", map[string]any{
		"Email":            email,
		"OTP":              otp,
		"ExpiresInMinutes": ttlMinutes(s.cfg.EmailVerifyOTPTTL),
		"AppName":          "Shortix",
		"Year":             s.now().UTC().Year(),
	})

	return &dto.MessageResponse{Message: "signup successful. verify email with otp"}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req dto.VerifyEmailRequest) (*dto.MessageResponse, error) {
	email := normalizeEmail(req.Email)
	storedOTP, err := s.otpStore.GetEmailVerificationOTP(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.ErrInvalidOTP
		}
		s.logger.Error("load email verify otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if !constantTimeEqual(storedOTP, req.OTP) {
		return nil, apperrors.ErrEmailVerificationCode
	}

	if err := s.users.MarkEmailVerified(ctx, email, s.now().UTC()); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		s.logger.Error("mark email verified failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	_ = s.otpStore.DeleteEmailVerificationOTP(ctx, email)
	return &dto.MessageResponse{Message: "email verified"}, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, meta RequestMeta) (*dto.AuthTokensResponse, error) {
	email := normalizeEmail(req.Email)
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrInvalidCredentials
		}
		s.logger.Error("load user for login failed", "error", err, "email", email)
		return nil, apperrors.InternalServerError()
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, apperrors.ErrUserInactive
	}
	if !user.IsEmailVerified {
		return nil, apperrors.ErrEmailNotVerified
	}

	now := s.now().UTC()
	sessionID := uuid.NewString()
	accessExpiresAt := now.Add(s.cfg.AccessTokenTTL)
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, user.Role, sessionID, accessExpiresAt)
	if err != nil {
		s.logger.Error("generate access token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	refreshToken, err := generateSecureToken(48)
	if err != nil {
		s.logger.Error("generate refresh token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	session := &model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		AccessTokenHash:  hashToken(accessToken),
		RefreshTokenHash: hashToken(refreshToken),
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		IPAddress:        pointerOrNil(meta.IP),
		UserAgent:        pointerOrNil(meta.UserAgent),
		Device:           pointerOrNil(meta.Device),
	}

	if err := s.sessions.CreateWithLastLogin(ctx, session, now, meta.IP, meta.UserAgent, meta.Device); err != nil {
		s.logger.Error("create session with login metadata failed", "error", err, "user_id", user.ID)
		return nil, apperrors.InternalServerError()
	}

	user.LastLoginAt = &now
	user.LastLoginIP = pointerOrNil(meta.IP)
	user.LastLoginUserAgent = pointerOrNil(meta.UserAgent)
	user.LastLoginDevice = pointerOrNil(meta.Device)

	s.sendEmailBestEffort(ctx, user.Email, "New login to your Shortix account", "login_alert.html", map[string]any{
		"Email":     user.Email,
		"LoggedAt":  now.Format(time.RFC1123Z),
		"IP":        valueOrUnknown(meta.IP),
		"Device":    valueOrUnknown(meta.Device),
		"UserAgent": valueOrUnknown(meta.UserAgent),
		"AppName":   "Shortix",
		"Year":      now.Year(),
	})

	return &dto.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest, meta RequestMeta) (*dto.AuthTokensResponse, error) {
	refreshHash := hashToken(req.RefreshToken)
	session, err := s.sessions.GetByRefreshHash(ctx, refreshHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrRefreshTokenInvalid
		}
		s.logger.Error("load session by refresh hash failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if session.IsRevoked {
		return nil, apperrors.ErrRefreshTokenInvalid
	}
	if s.now().UTC().After(session.RefreshExpiresAt) {
		return nil, apperrors.ErrRefreshTokenExpired
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		s.logger.Error("load user for refresh failed", "error", err, "user_id", session.UserID)
		return nil, apperrors.InternalServerError()
	}
	if !user.IsActive {
		return nil, apperrors.ErrUserInactive
	}

	if err := s.sessions.RevokeByRefreshHash(ctx, refreshHash); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("revoke old refresh session failed", "error", err)
			return nil, apperrors.InternalServerError()
		}
	}

	now := s.now().UTC()
	sessionID := uuid.NewString()
	accessExpiresAt := now.Add(s.cfg.AccessTokenTTL)
	refreshExpiresAt := now.Add(s.cfg.RefreshTokenTTL)

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID, user.Role, sessionID, accessExpiresAt)
	if err != nil {
		s.logger.Error("generate refreshed access token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	refreshToken, err := generateSecureToken(48)
	if err != nil {
		s.logger.Error("generate refreshed refresh token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	newSession := &model.Session{
		ID:               sessionID,
		UserID:           user.ID,
		AccessTokenHash:  hashToken(accessToken),
		RefreshTokenHash: hashToken(refreshToken),
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		IPAddress:        pointerOrNil(meta.IP),
		UserAgent:        pointerOrNil(meta.UserAgent),
		Device:           pointerOrNil(meta.Device),
	}
	if err := s.sessions.Create(ctx, newSession); err != nil {
		s.logger.Error("create refreshed session failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	return &dto.AuthTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toUserResponse(user),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID, sessionID string, req dto.LogoutRequest) (*dto.MessageResponse, error) {
	revoked := false
	if sessionID != "" {
		if err := s.sessions.RevokeByID(ctx, userID, sessionID); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				s.logger.Error("revoke session by id failed", "error", err)
				return nil, apperrors.InternalServerError()
			}
		} else {
			revoked = true
		}
	}

	if strings.TrimSpace(req.RefreshToken) != "" {
		if err := s.sessions.RevokeByRefreshHash(ctx, hashToken(req.RefreshToken)); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				s.logger.Error("revoke session by refresh hash failed", "error", err)
				return nil, apperrors.InternalServerError()
			}
		} else {
			revoked = true
		}
	}

	if revoked {
		s.sendSessionRevokedEmail(ctx, userID, "You logged out from one of your sessions.")
	}

	return &dto.MessageResponse{Message: "logged out"}, nil
}

func (s *AuthService) ListSessions(ctx context.Context, userID string) (*dto.SessionsResponse, error) {
	sessions, err := s.sessions.ListActiveByUser(ctx, userID)
	if err != nil {
		s.logger.Error("list sessions failed", "error", err, "user_id", userID)
		return nil, apperrors.InternalServerError()
	}

	resp := dto.SessionsResponse{Sessions: make([]dto.SessionResponse, 0, len(sessions))}
	for _, session := range sessions {
		resp.Sessions = append(resp.Sessions, dto.SessionResponse{
			ID:               session.ID,
			Device:           valueOrEmpty(session.Device),
			IP:               valueOrEmpty(session.IPAddress),
			UserAgent:        valueOrEmpty(session.UserAgent),
			CreatedAt:        session.CreatedAt,
			AccessExpiresAt:  session.AccessExpiresAt,
			RefreshExpiresAt: session.RefreshExpiresAt,
		})
	}
	return &resp, nil
}

func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID string) (*dto.MessageResponse, error) {
	if err := s.sessions.RevokeByID(ctx, userID, sessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrSessionNotFound
		}
		s.logger.Error("revoke user session failed", "error", err, "user_id", userID, "session_id", sessionID)
		return nil, apperrors.InternalServerError()
	}

	s.sendSessionRevokedEmail(ctx, userID, "A session was revoked from your account.")
	return &dto.MessageResponse{Message: "session revoked"}, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req dto.ForgotPasswordRequest) (*dto.MessageResponse, error) {
	email := normalizeEmail(req.Email)
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("load user for forgot password failed", "error", err)
			return nil, apperrors.InternalServerError()
		}
		return &dto.MessageResponse{Message: "if account exists, an otp has been sent"}, nil
	}

	otp, err := generateOTP(6)
	if err != nil {
		s.logger.Error("generate forgot password otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if err := s.otpStore.SetPasswordResetOTP(ctx, email, otp, s.cfg.PasswordResetOTPTTL); err != nil {
		s.logger.Error("store forgot password otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	s.sendEmailBestEffort(ctx, email, "Shortix password reset code", "password_reset_otp.html", map[string]any{
		"Email":            email,
		"OTP":              otp,
		"ExpiresInMinutes": ttlMinutes(s.cfg.PasswordResetOTPTTL),
		"AppName":          "Shortix",
		"Year":             s.now().UTC().Year(),
	})

	s.logger.Info("forgot password otp generated", "user_id", user.ID, "email", email)
	return &dto.MessageResponse{Message: "if account exists, an otp has been sent"}, nil
}

func (s *AuthService) VerifyForgotPasswordOTP(ctx context.Context, req dto.ForgotPasswordVerifyOTPRequest) (*dto.ForgotPasswordVerifyOTPResponse, error) {
	email := normalizeEmail(req.Email)
	storedOTP, err := s.otpStore.GetPasswordResetOTP(ctx, email)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.ErrInvalidOTP
		}
		s.logger.Error("load forgot password otp failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if !constantTimeEqual(storedOTP, req.OTP) {
		return nil, apperrors.ErrInvalidOTP
	}

	tempToken, err := generateSecureToken(48)
	if err != nil {
		s.logger.Error("generate temp token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}
	tempHash := hashToken(tempToken)
	if err := s.otpStore.SetPasswordResetTempToken(ctx, tempHash, email, s.cfg.PasswordResetTempTTL); err != nil {
		s.logger.Error("store temp token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	_ = s.otpStore.DeletePasswordResetOTP(ctx, email)
	return &dto.ForgotPasswordVerifyOTPResponse{TempToken: tempToken}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req dto.ResetPasswordRequest) (*dto.MessageResponse, error) {
	email := normalizeEmail(req.Email)
	tempTokenHash := hashToken(req.TempToken)
	storedEmail, err := s.otpStore.GetPasswordResetTempTokenEmail(ctx, tempTokenHash)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, apperrors.ErrInvalidTempToken
		}
		s.logger.Error("load temp token failed", "error", err)
		return nil, apperrors.InternalServerError()
	}
	if normalizeEmail(storedEmail) != email {
		return nil, apperrors.ErrInvalidTempToken
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		s.logger.Error("load user for reset password failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.cfg.BcryptCost)
	if err != nil {
		s.logger.Error("hash reset password failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if err := s.users.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		s.logger.Error("update password failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	if err := s.sessions.RevokeByUser(ctx, user.ID); err != nil {
		s.logger.Error("revoke user sessions failed", "error", err)
		return nil, apperrors.InternalServerError()
	}

	_ = s.otpStore.DeletePasswordResetTempToken(ctx, tempTokenHash)
	s.sendSessionRevokedEmail(ctx, user.ID, "Your password was reset and all active sessions were revoked.")
	return &dto.MessageResponse{Message: "password reset successful"}, nil
}

func toUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Role:            user.Role,
		IsActive:        user.IsActive,
		IsEmailVerified: user.IsEmailVerified,
		LastLoginAt:     user.LastLoginAt,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateOTP(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	builder := strings.Builder{}
	builder.Grow(length)
	for _, b := range bytes {
		builder.WriteByte('0' + (b % 10))
	}
	return builder.String(), nil
}

func generateSecureToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func pointerOrNil(v string) *string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func valueOrUnknown(v string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return "Unknown"
	}
	return trimmed
}

func ttlMinutes(ttl time.Duration) int {
	minutes := int(ttl / time.Minute)
	if minutes <= 0 {
		return 1
	}
	return minutes
}

func (s *AuthService) sendSessionRevokedEmail(ctx context.Context, userID, reason string) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error("load user for revoke email failed", "error", err, "user_id", userID)
		}
		return
	}

	now := s.now().UTC()
	s.sendEmailBestEffort(ctx, user.Email, "Security alert: session revoked", "session_revoked.html", map[string]any{
		"Email":     user.Email,
		"Reason":    reason,
		"RevokedAt": now.Format(time.RFC1123Z),
		"AppName":   "Shortix",
		"Year":      now.Year(),
	})
}

func (s *AuthService) sendEmailBestEffort(ctx context.Context, to, subject, templateName string, data map[string]any) {
	if strings.TrimSpace(to) == "" {
		return
	}
	if err := s.emails.SendTemplate(ctx, to, subject, templateName, data); err != nil {
		s.logger.Error("send email failed", "error", err, "template", templateName, "recipient", to)
	}
}
