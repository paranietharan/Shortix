package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shortix-api/internal/config"
	"shortix-api/internal/handler"
	"shortix-api/internal/middleware"
	"shortix-api/internal/repository"
	"shortix-api/internal/router"
	"shortix-api/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	userRepo := repository.NewPostgresUserRepository(db)
	sessionRepo := repository.NewPostgresSessionRepository(db)
	otpRepo := repository.NewRedisOTPRepository(redisClient)

	tokenManager := service.NewTokenManager(cfg.JWTSecret)
	emailSender := service.NewSMTPSender(cfg, logger)
	authService := service.NewAuthService(userRepo, sessionRepo, otpRepo, emailSender, tokenManager, cfg, logger)
	authHandler := handler.NewAuthHandler(authService)
	authMW := middleware.NewAuthMiddleware(tokenManager, sessionRepo)

	app := router.NewRouter(cfg, authHandler, authMW)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- app.Run(":" + cfg.Port)
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-serverErr:
		log.Fatalf("server stopped with error: %v", err)
	}
}
