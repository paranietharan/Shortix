package repository

import (
	"context"
	"errors"
	"time"

	"shortix-api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, userID string) (*model.User, error)
	MarkEmailVerified(ctx context.Context, email string, verifiedAt time.Time) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdatePartial(ctx context.Context, userID string, fields map[string]interface{}) error
	DeactivateUser(ctx context.Context, userID string) error
	ListUsers(ctx context.Context, page, limit int) ([]*model.User, int, error)
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	q := `
		INSERT INTO users (email, password_hash, role, is_active, is_email_verified)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, q, user.Email, user.PasswordHash, user.Role, user.IsActive, user.IsEmailVerified).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	q := `
		SELECT id, email, password_hash, role, is_active, is_email_verified, email_verified_at,
		       last_login_at, last_login_ip::text, last_login_user_agent, last_login_device,
		       first_name, last_name, profile_picture_url, bio, phone_number,
		       created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &model.User{}
	var lastLoginIP *string
	err := r.db.QueryRow(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.IsEmailVerified,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&lastLoginIP,
		&user.LastLoginUserAgent,
		&user.LastLoginDevice,
		&user.FirstName,
		&user.LastName,
		&user.ProfilePictureURL,
		&user.Bio,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	user.LastLoginIP = lastLoginIP
	return user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	q := `
		SELECT id, email, password_hash, role, is_active, is_email_verified, email_verified_at,
		       last_login_at, last_login_ip::text, last_login_user_agent, last_login_device,
		       first_name, last_name, profile_picture_url, bio, phone_number,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &model.User{}
	var lastLoginIP *string
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.IsEmailVerified,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&lastLoginIP,
		&user.LastLoginUserAgent,
		&user.LastLoginDevice,
		&user.FirstName,
		&user.LastName,
		&user.ProfilePictureURL,
		&user.Bio,
		&user.PhoneNumber,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	user.LastLoginIP = lastLoginIP
	return user, nil
}

func (r *PostgresUserRepository) MarkEmailVerified(ctx context.Context, email string, verifiedAt time.Time) error {
	q := `
		UPDATE users
		SET is_email_verified = TRUE,
		    email_verified_at = $2,
		    updated_at = NOW()
		WHERE email = $1
	`
	cmd, err := r.db.Exec(ctx, q, email, verifiedAt)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	q := `
		UPDATE users
		SET password_hash = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, q, userID, passwordHash)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) UpdateEmail(ctx context.Context, userID, email string) error {
	q := `
		UPDATE users
		SET email = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, q, userID, email)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) UpdatePartial(ctx context.Context, userID string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	query := "UPDATE users SET "
	args := []interface{}{}
	argID := 1

	for k, v := range fields {
		query += k + " = $" + string(rune('0'+argID)) + ", "
		args = append(args, v)
		argID++
	}

	query += "updated_at = NOW() WHERE id = $" + string(rune('0'+argID))
	args = append(args, userID)

	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) DeactivateUser(ctx context.Context, userID string) error {
	q := `UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1`
	cmd, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresUserRepository) ListUsers(ctx context.Context, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	q := `
		SELECT id, email, role, is_active, is_email_verified, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsActive, &u.IsEmailVerified, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}
