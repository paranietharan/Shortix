package repository

import (
	"context"
	"shortix-api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByShortCode(ctx context.Context, code string) (*model.URL, error)
	GetByCustomAlias(ctx context.Context, alias string) (*model.URL, error)
	GetByLongURLAndUser(ctx context.Context, longURL string, userID string) (*model.URL, error)
	GetByID(ctx context.Context, id string) (*model.URL, error)
	Delete(ctx context.Context, id string) error
	IsAliasAvailable(ctx context.Context, alias string) (bool, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.URL, int64, error)
	ExistsByCustomAlias(ctx context.Context, alias string) (bool, error)
}

type urlRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) URLRepository {
	return &urlRepository{
		db: db,
	}
}

func (r *urlRepository) Create(ctx context.Context, url *model.URL) error {
	query := `
	INSERT INTO urls (user_id, long_url, short_code, custom_alias, expires_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at;
	`

	return r.db.QueryRow(ctx,
		query,
		url.UserID,
		url.LongURL,
		url.ShortCode,
		url.CustomAlias,
		url.ExpiresAt,
	).Scan(&url.ID, &url.CreatedAt)
}

func (r *urlRepository) GetByShortCode(ctx context.Context, code string) (*model.URL, error) {
	query := `
	SELECT id, user_id, long_url, short_code, custom_alias, expires_at, created_at
	FROM urls
	WHERE short_code = $1;
	`

	var url model.URL
	err := r.db.QueryRow(ctx, query, code).Scan(
		&url.ID,
		&url.UserID,
		&url.LongURL,
		&url.ShortCode,
		&url.CustomAlias,
		&url.ExpiresAt,
		&url.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *urlRepository) GetByCustomAlias(ctx context.Context, alias string) (*model.URL, error) {
	query := `
	SELECT id, user_id, long_url, short_code, custom_alias, expires_at, created_at
	FROM urls
	WHERE custom_alias = $1;
	`

	var url model.URL
	err := r.db.QueryRow(ctx, query, alias).Scan(
		&url.ID,
		&url.UserID,
		&url.LongURL,
		&url.ShortCode,
		&url.CustomAlias,
		&url.ExpiresAt,
		&url.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *urlRepository) GetByLongURLAndUser(ctx context.Context, longURL string, userID string) (*model.URL, error) {
	query := `
	SELECT id, user_id, long_url, short_code, custom_alias, expires_at, created_at
	FROM urls
	WHERE long_url = $1 AND user_id = $2;
	`

	var url model.URL
	err := r.db.QueryRow(ctx, query, longURL, userID).Scan(
		&url.ID,
		&url.UserID,
		&url.LongURL,
		&url.ShortCode,
		&url.CustomAlias,
		&url.ExpiresAt,
		&url.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *urlRepository) GetByID(ctx context.Context, id string) (*model.URL, error) {
	query := `
	SELECT id, user_id, long_url, short_code, custom_alias, expires_at, created_at
	FROM urls
	WHERE id = $1;
	`

	var url model.URL
	err := r.db.QueryRow(ctx, query, id).Scan(
		&url.ID,
		&url.UserID,
		&url.LongURL,
		&url.ShortCode,
		&url.CustomAlias,
		&url.ExpiresAt,
		&url.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &url, nil
}

func (r *urlRepository) IsAliasAvailable(ctx context.Context, alias string) (bool, error) {
	query := `SELECT EXISTS (
		SELECT 1 FROM urls WHERE custom_alias = $1 OR short_code = $1
	);`

	var exists bool
	err := r.db.QueryRow(ctx, query, alias).Scan(&exists)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func (r *urlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM urls WHERE id = $1;`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *urlRepository) ListByUser(ctx context.Context, userID string, page, limit int) ([]*model.URL, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM urls WHERE user_id = $1`
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
	SELECT id, user_id, long_url, short_code, custom_alias, expires_at, created_at
	FROM urls
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3;
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var urls []*model.URL
	for rows.Next() {
		var url model.URL
		err := rows.Scan(
			&url.ID,
			&url.UserID,
			&url.LongURL,
			&url.ShortCode,
			&url.CustomAlias,
			&url.ExpiresAt,
			&url.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		urls = append(urls, &url)
	}

	return urls, total, nil
}

func (r *urlRepository) ExistsByCustomAlias(ctx context.Context, alias string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM urls WHERE custom_alias = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, alias).Scan(&exists)
	return exists, err
}
