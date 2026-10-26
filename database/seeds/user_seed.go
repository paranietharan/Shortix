package seeds

import (
	"context"
	"log"
	"shortix-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type userSeed struct{}

func NewUserSeed() Seed {
	return &userSeed{}
}

func (s *userSeed) Run(ctx context.Context, db *pgxpool.Pool, cfg *config.Config) error {
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		log.Println("skipping user seed: SEED_ADMIN_EMAIL or SEED_ADMIN_PASSWORD not set")
		return nil
	}

	// Check if user already exists
	var exists bool
	err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", cfg.SeedAdminEmail).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Printf("user seed skipped: user %s already exists\n", cfg.SeedAdminEmail)
		return nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.SeedAdminPassword), cfg.BcryptCost)
	if err != nil {
		return err
	}

	// Insert admin user
	query := `
		INSERT INTO users (email, password_hash, role, is_email_verified)
		VALUES ($1, $2, 'ADMIN', TRUE)
	`
	_, err = db.Exec(ctx, query, cfg.SeedAdminEmail, string(hashedPassword))
	if err != nil {
		return err
	}

	log.Printf("user seed executed: created user %s with ADMIN role\n", cfg.SeedAdminEmail)
	return nil
}
