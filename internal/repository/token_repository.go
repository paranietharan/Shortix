package repository

import (
	"context"
	"errors"
	"time"

	"shortix-api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository interface {
	CreateWithLastLogin(ctx context.Context, session *model.Session, loginAt time.Time, ip, userAgent, device string) error
	Create(ctx context.Context, session *model.Session) error
	GetByRefreshHash(ctx context.Context, refreshHash string) (*model.Session, error)
	GetByAccessHash(ctx context.Context, accessHash string) (*model.Session, error)
	ListActiveByUser(ctx context.Context, userID string) ([]model.Session, error)
	RevokeByID(ctx context.Context, userID, sessionID string) error
	RevokeByRefreshHash(ctx context.Context, refreshHash string) error
	RevokeByUser(ctx context.Context, userID string) error
}

type PostgresSessionRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSessionRepository(db *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) CreateWithLastLogin(ctx context.Context, session *model.Session, loginAt time.Time, ip, userAgent, device string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertSession := `
		INSERT INTO sessions (
			id, user_id, access_token_hash, refresh_token_hash,
			access_expires_at, refresh_expires_at, ip_address, user_agent, device, is_revoked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8, $9, FALSE)
		RETURNING created_at, updated_at
	`

	if err := tx.QueryRow(ctx, insertSession,
		session.ID,
		session.UserID,
		session.AccessTokenHash,
		session.RefreshTokenHash,
		session.AccessExpiresAt,
		session.RefreshExpiresAt,
		nilIfEmpty(ip),
		nilIfEmpty(userAgent),
		nilIfEmpty(device),
	).Scan(&session.CreatedAt, &session.UpdatedAt); err != nil {
		return err
	}

	updateUser := `
		UPDATE users
		SET last_login_at = $2,
		    last_login_ip = NULLIF($3, '')::inet,
		    last_login_user_agent = NULLIF($4, ''),
		    last_login_device = NULLIF($5, ''),
		    updated_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.Exec(ctx, updateUser, session.UserID, loginAt, ip, userAgent, device); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresSessionRepository) Create(ctx context.Context, session *model.Session) error {
	q := `
		INSERT INTO sessions (
			id, user_id, access_token_hash, refresh_token_hash,
			access_expires_at, refresh_expires_at, ip_address, user_agent, device, is_revoked
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8, $9, FALSE)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(ctx, q,
		session.ID,
		session.UserID,
		session.AccessTokenHash,
		session.RefreshTokenHash,
		session.AccessExpiresAt,
		session.RefreshExpiresAt,
		nilIfEmpty(ptrValue(session.IPAddress)),
		nilIfEmpty(ptrValue(session.UserAgent)),
		nilIfEmpty(ptrValue(session.Device)),
	).Scan(&session.CreatedAt, &session.UpdatedAt)
}

func (r *PostgresSessionRepository) GetByRefreshHash(ctx context.Context, refreshHash string) (*model.Session, error) {
	q := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       access_expires_at, refresh_expires_at, is_revoked,
		       ip_address::text, user_agent, device, created_at, updated_at
		FROM sessions
		WHERE refresh_token_hash = $1
	`
	session := &model.Session{}
	err := r.db.QueryRow(ctx, q, refreshHash).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AccessExpiresAt,
		&session.RefreshExpiresAt,
		&session.IsRevoked,
		&session.IPAddress,
		&session.UserAgent,
		&session.Device,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return session, nil
}

func (r *PostgresSessionRepository) GetByAccessHash(ctx context.Context, accessHash string) (*model.Session, error) {
	q := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       access_expires_at, refresh_expires_at, is_revoked,
		       ip_address::text, user_agent, device, created_at, updated_at
		FROM sessions
		WHERE access_token_hash = $1
	`
	session := &model.Session{}
	err := r.db.QueryRow(ctx, q, accessHash).Scan(
		&session.ID,
		&session.UserID,
		&session.AccessTokenHash,
		&session.RefreshTokenHash,
		&session.AccessExpiresAt,
		&session.RefreshExpiresAt,
		&session.IsRevoked,
		&session.IPAddress,
		&session.UserAgent,
		&session.Device,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return session, nil
}

func (r *PostgresSessionRepository) ListActiveByUser(ctx context.Context, userID string) ([]model.Session, error) {
	q := `
		SELECT id, user_id, access_token_hash, refresh_token_hash,
		       access_expires_at, refresh_expires_at, is_revoked,
		       ip_address::text, user_agent, device, created_at, updated_at
		FROM sessions
		WHERE user_id = $1
		  AND is_revoked = FALSE
		  AND refresh_expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]model.Session, 0)
	for rows.Next() {
		var s model.Session
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.AccessTokenHash,
			&s.RefreshTokenHash,
			&s.AccessExpiresAt,
			&s.RefreshExpiresAt,
			&s.IsRevoked,
			&s.IPAddress,
			&s.UserAgent,
			&s.Device,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return sessions, nil
}

func (r *PostgresSessionRepository) RevokeByID(ctx context.Context, userID, sessionID string) error {
	q := `
		UPDATE sessions
		SET is_revoked = TRUE,
		    updated_at = NOW()
		WHERE id = $1
		  AND user_id = $2
		  AND is_revoked = FALSE
	`
	cmd, err := r.db.Exec(ctx, q, sessionID, userID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresSessionRepository) RevokeByRefreshHash(ctx context.Context, refreshHash string) error {
	q := `
		UPDATE sessions
		SET is_revoked = TRUE,
		    updated_at = NOW()
		WHERE refresh_token_hash = $1
	`
	cmd, err := r.db.Exec(ctx, q, refreshHash)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *PostgresSessionRepository) RevokeByUser(ctx context.Context, userID string) error {
	q := `
		UPDATE sessions
		SET is_revoked = TRUE,
		    updated_at = NOW()
		WHERE user_id = $1
		  AND is_revoked = FALSE
	`
	_, err := r.db.Exec(ctx, q, userID)
	return err
}

func ptrValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func nilIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
