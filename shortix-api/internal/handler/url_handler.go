package handler

import (
	"net/http"
	"shortix-api/internal/dto"
	"shortix-api/internal/middleware"
	"shortix-api/internal/model"
	"shortix-api/internal/service"
	"shortix-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(urlService service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

func (h *URLHandler) CreateURL(c *gin.Context) {
	// Get userID from auth middleware
	userID, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		response.Error(c.Writer, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c.Writer, http.StatusBadRequest, "Invalid request")
		return
	}

	res, isNew, err := h.urlService.CreateURL(c.Request.Context(), userID.(string), &req)
	if err != nil {
		response.Error(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	message := "URL shortened successfully"
	if !isNew {
		message = "URL already exists"
	}

	response.Success(c.Writer, http.StatusCreated, message, res)
}

func (h *URLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("short_code")

	// Collect analytics data
	clickData := &model.Click{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referrer:  c.Request.Referer(),
		Device:    service.ParseDeviceFromUserAgent(c.Request.UserAgent()),
	}

	longURL, err := h.urlService.GetRedirectURL(c.Request.Context(), shortCode, clickData)
	if err != nil {
		// Handle errors (link not found, expired, etc.)
		if err.Error() == "link not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
			return
		}
		if err.Error() == "link expired" {
			c.JSON(http.StatusGone, gin.H{"error": "Link expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Use 302 for analytics accuracy (301 might be cached by browsers)
	c.Redirect(http.StatusFound, longURL)
}

func (h *URLHandler) GetAnalytics(c *gin.Context) {
	urlID := c.Param("id")

	res, err := h.urlService.GetAnalytics(c.Request.Context(), urlID)
	if err != nil {
		response.Error(c.Writer, http.StatusInternalServerError, "Failed to fetch analytics")
		return
	}

	response.Success(c.Writer, http.StatusOK, "Analytics fetched successfully", res)
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	urlID := c.Param("id")
	userID, _ := c.Get(middleware.ContextUserIDKey)
	role, _ := c.Get(middleware.ContextRoleKey)

	err := h.urlService.DeleteURL(c.Request.Context(), urlID, userID.(string), role.(string))
	if err != nil {
		if err.Error() == "link not found" {
			response.Error(c.Writer, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "permission denied" {
			response.Error(c.Writer, http.StatusForbidden, err.Error())
			return
		}
		response.Error(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c.Writer, http.StatusOK, "URL deleted successfully", nil)
}
