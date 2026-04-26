package handler

import (
	"net/http"

	"shortix-api/internal/dto"
	apperrors "shortix-api/internal/errors"
	"shortix-api/internal/middleware"
	"shortix-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.Signup(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.VerifyEmail(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	meta := requestMetaFromContext(c)
	resp, err := h.svc.Login(c.Request.Context(), req, meta)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	meta := requestMetaFromContext(c)
	resp, err := h.svc.Refresh(c.Request.Context(), req, meta)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			h.writeError(c, apperrors.ErrValidation)
			return
		}
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}
	sessionID, _ := middleware.SessionIDFromContext(c)

	resp, err := h.svc.Logout(c.Request.Context(), userID, sessionID, req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.ListSessions(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		h.writeError(c, apperrors.ErrBadRequest)
		return
	}

	resp, err := h.svc.RevokeSession(c.Request.Context(), userID, sessionID)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.ForgotPassword(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyForgotPasswordOTP(c *gin.Context) {
	var req dto.ForgotPasswordVerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.VerifyForgotPasswordOTP(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.ResetPassword(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RequestEmailChange(c *gin.Context) {
	var req dto.EmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.RequestEmailChange(c.Request.Context(), userID, req.NewEmail)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyEmailChange(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.VerifyEmailChange(c.Request.Context(), userID, req.OTP)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RequestPasswordChange(c *gin.Context) {
	var req dto.PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.RequestPasswordChange(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) VerifyPasswordChange(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.VerifyPasswordChange(c.Request.Context(), userID, req.OTP)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	if err := h.svc.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "profile updated successfully"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		h.writeError(c, apperrors.ErrUnauthorized)
		return
	}

	resp, err := h.svc.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) AdminDeactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.writeError(c, apperrors.ErrBadRequest)
		return
	}

	if err := h.svc.AdminDeactivateUser(c.Request.Context(), userID); err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "user deactivated successfully"})
}

func (h *AuthHandler) AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		h.writeError(c, apperrors.ErrBadRequest)
		return
	}

	if err := h.svc.AdminDeactivateUser(c.Request.Context(), userID); err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "user soft deleted successfully"})
}

func (h *AuthHandler) AdminListUsers(c *gin.Context) {
	var query struct {
		Page  int `form:"page,default=1" binding:"omitempty,min=1"`
		Limit int `form:"limit,default=20" binding:"omitempty,min=1,max=100"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		h.writeError(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.svc.AdminListUsers(c.Request.Context(), query.Page, query.Limit)
	if err != nil {
		h.writeError(c, apperrors.AsAppError(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) writeError(c *gin.Context, appErr *apperrors.AppError) {
	c.JSON(appErr.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
}

func requestMetaFromContext(c *gin.Context) service.RequestMeta {
	ua := c.GetHeader("User-Agent")
	return service.RequestMeta{
		IP:        c.ClientIP(),
		UserAgent: ua,
		Device:    service.ParseDeviceFromUserAgent(ua),
	}
}
