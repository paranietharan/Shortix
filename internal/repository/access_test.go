package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
)

func TestAnalyticsBelongsToOwnerAndExpiredLinksAreNotReused(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL required")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 1
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(ctx, `CREATE TEMP TABLE urls(id uuid,user_id uuid,long_url text,short_code text,custom_alias text,expires_at timestamptz,created_at timestamptz);
        CREATE TEMP TABLE clicks(url_id uuid,clicked_at timestamptz,referrer text,device text)`)
	if err != nil {
		t.Fatal(err)
	}
	id, owner := uuid.NewString(), uuid.NewString()
	if _, err = db.Exec(ctx, "INSERT INTO urls VALUES($1,$2,'https://example.com','abc',NULL,now()-interval '1 day',now())", id, owner); err != nil {
		t.Fatal(err)
	}
	repo := NewClickRepository(db)
	if _, err = repo.GetAnalytics(ctx, id, owner); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.GetAnalytics(ctx, id, uuid.NewString()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("foreign analytics: %v", err)
	}
	link, err := NewURLRepository(db).GetByLongURLAndUser(ctx, "https://example.com", owner)
	if link != nil || (err != nil && !errors.Is(err, pgx.ErrNoRows)) {
		t.Fatalf("expired link reused: %v %v", link, err)
	}
}
