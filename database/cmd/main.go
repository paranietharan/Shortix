package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	migrate "shortix-api/database/runner"
	"shortix-api/internal/config"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	migrateAction := flag.String("migrate", "", "migration action: up|down")
	seed := flag.Bool("seed", false, "seed default admin and user from config/env")
	forceVersion := flag.Int("version", 0, "migration version to force when using -migrate force")
	flag.Parse()

	// Load environment variables from .env files
	root, err := projectRoot()
	if err != nil {
		log.Fatalf("failed to read working directory: %v", err)
	}
	loadDotEnv(filepath.Join(root, "database", ".env"))
	loadDotEnv(filepath.Join(root, ".env"))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	if *migrateAction != "" {
		if err := runMigrate(dsn, *migrateAction, *forceVersion, root); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		return
	}

	if *seed {
		if err := runSeed(dsn); err != nil {
			log.Fatalf("seed failed: %v", err)
		}
		return
	}

	log.Println("no action specified, exiting")

}

func runSeed(dsn string) error {
	cfg := config.Load()

	adminEmail := normalizeEmail(cfg.SeedAdminEmail)
	adminPassword := strings.TrimSpace(cfg.SeedAdminPassword)
	userEmail := normalizeEmail(cfg.SeedUserEmail)
	userPassword := strings.TrimSpace(cfg.SeedUserPassword)

	if adminEmail == "" || adminPassword == "" || userEmail == "" || userPassword == "" {
		return fmt.Errorf("missing seed config values: set SEED_ADMIN_EMAIL, SEED_ADMIN_PASSWORD, SEED_USER_EMAIL, and SEED_USER_PASSWORD")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect db for seed: %w", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping db for seed: %w", err)
	}

	if err := upsertSeedUser(ctx, db, adminEmail, adminPassword, "ADMIN", cfg.BcryptCost); err != nil {
		return fmt.Errorf("seed admin user: %w", err)
	}

	if err := upsertSeedUser(ctx, db, userEmail, userPassword, "USER", cfg.BcryptCost); err != nil {
		return fmt.Errorf("seed regular user: %w", err)
	}

	log.Println("seed completed: admin and user upserted")
	return nil
}

func upsertSeedUser(ctx context.Context, db *pgxpool.Pool, email, password, role string, bcryptCost int) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password for %s: %w", email, err)
	}

	q := `
		INSERT INTO users (email, password_hash, role, is_active, is_email_verified, email_verified_at)
		VALUES ($1, $2, $3, TRUE, TRUE, NOW())
		ON CONFLICT (email)
		DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			is_active = TRUE,
			is_email_verified = TRUE,
			email_verified_at = COALESCE(users.email_verified_at, NOW()),
			updated_at = NOW()
	`

	if _, err := db.Exec(ctx, q, email, string(hash), role); err != nil {
		return err
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func runMigrate(dsn, action string, version int, root string) error {
	if err := prepareLegacySchemaMigrations(dsn); err != nil {
		return err
	}

	dir := filepath.Join(root, "database", "migrations")
	r := migrate.MigrationRunner{SourceDir: "file://" + dir, DBURL: dsn}

	switch action {
	case "up":
		return r.Up()
	case "down":
		return r.Down()
	case "force":
		return r.Force(version)
	default:
		return fmt.Errorf("unsupported migrate action %q (use up, down, or force)", action)
	}
}

func prepareLegacySchemaMigrations(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect db for migration metadata: %w", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping db for migration metadata: %w", err)
	}

	var hasAppliedAt bool
	err = db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'schema_migrations'
			  AND column_name = 'applied_at'
		)
	`).Scan(&hasAppliedAt)
	if err != nil {
		return fmt.Errorf("check schema_migrations shape: %w", err)
	}

	if !hasAppliedAt {
		return nil
	}

	log.Println("detected legacy schema_migrations table, renaming to schema_migrations_legacy")
	_, err = db.Exec(ctx, `ALTER TABLE IF EXISTS schema_migrations RENAME TO schema_migrations_legacy`)
	if err != nil {
		return fmt.Errorf("rename legacy schema_migrations table: %w", err)
	}

	return nil
}

func projectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if filepath.Base(cwd) == "database" {
		return filepath.Dir(cwd), nil
	}
	if filepath.Base(filepath.Dir(cwd)) == "database" {
		return filepath.Dir(filepath.Dir(cwd)), nil
	}
	return cwd, nil
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("warning: failed to scan %s: %v", path, err)
	}
}
