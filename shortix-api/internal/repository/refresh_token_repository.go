package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"shortix-api/internal/model"

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
	SetEmailChangeData(ctx context.Context, userID string, data *model.EmailChangeData, ttl time.Duration) error
	GetEmailChangeData(ctx context.Context, userID string) (*model.EmailChangeData, error)
	DeleteEmailChangeData(ctx context.Context, userID string) error
	SetPasswordChangeData(ctx context.Context, userID string, data *model.PasswordChangeData, ttl time.Duration) error
	GetPasswordChangeData(ctx context.Context, userID string) (*model.PasswordChangeData, error)
	DeletePasswordChangeData(ctx context.Context, userID string) error
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

func (r *RedisOTPRepository) SetEmailChangeData(ctx context.Context, userID string, data *model.EmailChangeData, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, emailChangeKey(userID), b, ttl).Err()
}

func (r *RedisOTPRepository) GetEmailChangeData(ctx context.Context, userID string) (*model.EmailChangeData, error) {
	b, err := r.client.Get(ctx, emailChangeKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var data model.EmailChangeData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *RedisOTPRepository) DeleteEmailChangeData(ctx context.Context, userID string) error {
	return r.client.Del(ctx, emailChangeKey(userID)).Err()
}

func (r *RedisOTPRepository) SetPasswordChangeData(ctx context.Context, userID string, data *model.PasswordChangeData, ttl time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, passwordChangeKey(userID), b, ttl).Err()
}

func (r *RedisOTPRepository) GetPasswordChangeData(ctx context.Context, userID string) (*model.PasswordChangeData, error) {
	b, err := r.client.Get(ctx, passwordChangeKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var data model.PasswordChangeData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *RedisOTPRepository) DeletePasswordChangeData(ctx context.Context, userID string) error {
	return r.client.Del(ctx, passwordChangeKey(userID)).Err()
}

func emailChangeKey(userID string) string {
	return fmt.Sprintf("email_change:%s", userID)
}

func passwordChangeKey(userID string) string {
	return fmt.Sprintf("password_change:%s", userID)
}
