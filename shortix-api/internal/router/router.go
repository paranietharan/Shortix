package router

import (
	"shortix-api/internal/config"
	"shortix-api/internal/handler"
	"shortix-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	urlHandler *handler.URLHandler,
	authMW *middleware.AuthMiddleware,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	loginLimiter := middleware.NewRateLimiter(cfg.RateLimitWindow, cfg.RateLimitLoginMax)
	forgotPasswordLimiter := middleware.NewRateLimiter(cfg.RateLimitWindow, cfg.RateLimitForgotPassMax)

	// Redirect path (Critical Path)
	r.GET("/:short_code", urlHandler.Redirect)

	auth := r.Group("/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/login", loginLimiter.Middleware("login"), authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/forgot-password", forgotPasswordLimiter.Middleware("forgot-password"), authHandler.ForgotPassword)
		auth.POST("/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
		auth.POST("/forgot-password/reset", authHandler.ResetPassword)

		authSecured := auth.Group("")
		authSecured.Use(authMW.RequireAuth())
		{
			authSecured.POST("/logout", authHandler.Logout)
			authSecured.GET("/sessions", authHandler.ListSessions)
			authSecured.DELETE("/sessions/:id", authHandler.RevokeSession)
		}
	}

	// URL Operations
	urls := r.Group("/urls")
	urls.Use(authMW.RequireAuth())
	{
		urls.POST("", urlHandler.CreateURL)
		urls.GET("/:id/analytics", urlHandler.GetAnalytics)
	}

	protected := r.Group("/")
	protected.Use(authMW.RequireAuth())
	{
		protected.GET("/owner-only", middleware.RequireRoles("OWNER"), authHandler.OwnerOnly)
		protected.GET("/admin-only", middleware.RequireRoles("ADMIN"), authHandler.AdminOnly)
		protected.GET("/admin-owner", middleware.RequireRoles("ADMIN", "OWNER"), authHandler.AdminOwner)
	}

	return r
}
