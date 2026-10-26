package runner

import (
	"context"
	"log"
	"shortix-api/database/seeds"
	"shortix-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeedRunner struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewSeedRunner(db *pgxpool.Pool, cfg *config.Config) *SeedRunner {
	return &SeedRunner{
		db:  db,
		cfg: cfg,
	}
}

func (r *SeedRunner) Run(ctx context.Context) error {
	log.Println("starting database seeding...")
	if err := seeds.Execute(ctx, r.db, r.cfg); err != nil {
		return err
	}
	log.Println("database seeding completed")
	return nil
}
