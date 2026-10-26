# Migrations and seeding steps

## Migrations
Apply migrations from scratch:
```bash
go run database/cmd/main.go -migrate up
```

Rollback the last applied migration:
```bash
go run database/cmd/main.go -migrate down
```

If a migration fails and leaves the database dirty, reset the version first, then apply again:
```bash
go run database/cmd/main.go -migrate force -version -1
go run database/cmd/main.go -migrate up
```

## Seed Users
Seed default admin and regular user from config/env values:
```bash
go run database/cmd/main.go -seed
```

Required env values (in `.env`):
```env
SEED_ADMIN_EMAIL=admin@shortix.com
SEED_ADMIN_PASSWORD=Admin@12345
SEED_USER_EMAIL=user@shortix.com
SEED_USER_PASSWORD=User@12345
```

Notes:
- Seeding is idempotent (upsert by email).
- Existing seeded users are updated with new password/role and marked active + verified.