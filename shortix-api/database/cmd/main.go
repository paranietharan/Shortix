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
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	migrateAction := flag.String("migrate", "", "migration action: up|down")
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
		if err := runMigrate(dsn, *migrateAction, root); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		return
	}

	log.Println("no action specified, exiting")

}

func runMigrate(dsn, action string, root string) error {
	if err := prepareLegacySchemaMigrations(dsn); err != nil {
		return err
	}

	dir := filepath.Join(root, "database", "migrations")
	r := migrate.Runner{SourceDir: "file://" + dir, DBURL: dsn}

	switch action {
	case "up":
		return r.Up()
	case "down":
		return r.Down()
	default:
		return fmt.Errorf("unsupported migrate action %q (use up or down)", action)
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
