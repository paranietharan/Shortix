package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                 string
	Port                   string
	DatabaseURL            string
	SeedAdminEmail         string
	SeedAdminPassword      string
	SeedUserEmail          string
	SeedUserPassword       string
	RedisAddr              string
	RedisPassword          string
	RedisDB                int
	SMTPHost               string
	SMTPPort               string
	SMTPUsername           string
	SMTPPassword           string
	SMTPFromEmail          string
	SMTPFromName           string
	JWTSecret              string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
	EmailVerifyOTPTTL      time.Duration
	PasswordResetOTPTTL    time.Duration
	PasswordResetTempTTL   time.Duration
	BcryptCost             int
	RateLimitWindow        time.Duration
	RateLimitLoginMax      int
	RateLimitForgotPassMax int
}

func Load() *Config {
	// Best-effort load of local .env file for development.
	_ = godotenv.Load()

	return &Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://postgres:root@localhost:5432/shortix?sslmode=disable"),
		SeedAdminEmail:         getEnv("SEED_ADMIN_EMAIL", ""),
		SeedAdminPassword:      getEnv("SEED_ADMIN_PASSWORD", ""),
		SeedUserEmail:          getEnv("SEED_USER_EMAIL", ""),
		SeedUserPassword:       getEnv("SEED_USER_PASSWORD", ""),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:          getEnv("REDIS_PASSWORD", ""),
		RedisDB:                getEnvInt("REDIS_DB", 0),
		SMTPHost:               getEnv("SMTP_HOST", ""),
		SMTPPort:               getEnv("SMTP_PORT", "587"),
		SMTPUsername:           getEnv("SMTP_USERNAME", ""),
		SMTPPassword:           getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail:          getEnv("SMTP_FROM_EMAIL", ""),
		SMTPFromName:           getEnv("SMTP_FROM_NAME", "Shortix"),
		JWTSecret:              getEnv("JWT_SECRET", "change-this-in-production"),
		AccessTokenTTL:         getEnvDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL:        getEnvDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		EmailVerifyOTPTTL:      getEnvDuration("EMAIL_VERIFY_OTP_TTL", 10*time.Minute),
		PasswordResetOTPTTL:    getEnvDuration("PASSWORD_RESET_OTP_TTL", 10*time.Minute),
		PasswordResetTempTTL:   getEnvDuration("PASSWORD_RESET_TEMP_TTL", 15*time.Minute),
		BcryptCost:             getEnvInt("BCRYPT_COST", 12),
		RateLimitWindow:        getEnvDuration("RATE_LIMIT_WINDOW", time.Minute),
		RateLimitLoginMax:      getEnvInt("RATE_LIMIT_LOGIN_MAX", 10),
		RateLimitForgotPassMax: getEnvInt("RATE_LIMIT_FORGOT_PASSWORD_MAX", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
