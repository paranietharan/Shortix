package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPRepository interface {
	SetEmailVerificationOTP(ctx context.Context, email, otp string, ttl time.Duration) error
	GetEmailVerificationOTP(ctx context.Context, email string) (string, error)
	DeleteEmailVerificationOTP(ctx context.Context, email string) error
	SetPasswordResetOTP(ctx context.Context, email, otp string, ttl time.Duration) error
	GetPasswordResetOTP(ctx context.Context, email string) (string, error)
	DeletePasswordResetOTP(ctx context.Context, email string) error
	SetPasswordResetTempToken(ctx context.Context, tokenHash, email string, ttl time.Duration) error
	GetPasswordResetTempTokenEmail(ctx context.Context, tokenHash string) (string, error)
	DeletePasswordResetTempToken(ctx context.Context, tokenHash string) error
}

type RedisOTPRepository struct {
	client *redis.Client
}

func NewRedisOTPRepository(client *redis.Client) *RedisOTPRepository {
	return &RedisOTPRepository{client: client}
}

func (r *RedisOTPRepository) SetEmailVerificationOTP(ctx context.Context, email, otp string, ttl time.Duration) error {
	return r.client.Set(ctx, emailVerificationKey(email), otp, ttl).Err()
}

func (r *RedisOTPRepository) GetEmailVerificationOTP(ctx context.Context, email string) (string, error) {
	return r.client.Get(ctx, emailVerificationKey(email)).Result()
}

func (r *RedisOTPRepository) DeleteEmailVerificationOTP(ctx context.Context, email string) error {
	return r.client.Del(ctx, emailVerificationKey(email)).Err()
}

func (r *RedisOTPRepository) SetPasswordResetOTP(ctx context.Context, email, otp string, ttl time.Duration) error {
	return r.client.Set(ctx, passwordResetOTPKey(email), otp, ttl).Err()
}

func (r *RedisOTPRepository) GetPasswordResetOTP(ctx context.Context, email string) (string, error) {
	return r.client.Get(ctx, passwordResetOTPKey(email)).Result()
}

func (r *RedisOTPRepository) DeletePasswordResetOTP(ctx context.Context, email string) error {
	return r.client.Del(ctx, passwordResetOTPKey(email)).Err()
}

func (r *RedisOTPRepository) SetPasswordResetTempToken(ctx context.Context, tokenHash, email string, ttl time.Duration) error {
	return r.client.Set(ctx, passwordResetTempTokenKey(tokenHash), email, ttl).Err()
}

func (r *RedisOTPRepository) GetPasswordResetTempTokenEmail(ctx context.Context, tokenHash string) (string, error) {
	return r.client.Get(ctx, passwordResetTempTokenKey(tokenHash)).Result()
}

func (r *RedisOTPRepository) DeletePasswordResetTempToken(ctx context.Context, tokenHash string) error {
	return r.client.Del(ctx, passwordResetTempTokenKey(tokenHash)).Err()
}

func emailVerificationKey(email string) string {
	return "auth:verify-email:otp:" + email
}

func passwordResetOTPKey(email string) string {
	return "auth:forgot-password:otp:" + email
}

func passwordResetTempTokenKey(tokenHash string) string {
	return "auth:forgot-password:temp-token:" + tokenHash
}
