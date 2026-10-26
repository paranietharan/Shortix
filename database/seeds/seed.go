package seeds

import (
	"context"
	"log"
	"shortix-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Seed interface {
	Run(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) error
}

func All() []Seed {
	return []Seed{
		NewUserSeed(),
	}
}

func Execute(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) error {
	for _, seed := range All() {
		if err := seed.Run(ctx, db, cfg); err != nil {
			return err
		}
	}
	log.Println("all seeds executed successfully")
	return nil
}
