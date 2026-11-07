# Shortix

A Go URL-shortening API with authenticated link management, custom aliases, expiring redirects, Redis caching and asynchronous click analytics. PostgreSQL stores users, links and click records; Mailpit supports local email verification flows.

## Run locally

```sh
cp .env.example .env
# Edit .env before starting: choose your own JWT secret and seed credentials.
docker compose up --build
```

The API listens on port 8080 by default; Mailpit is at http://localhost:8025. Compose configures migrations and development seeds on startup. Set `RUN_SEED=false` outside a disposable development setup. The sample database/Redis addresses target Compose service names; use localhost addresses if running the Go process directly.

For native development, install Go 1.27.3, start the supporting services, update `.env`, and run `go run ./cmd/api`.

## API overview

| Route | Purpose |
| --- | --- |
| `POST /auth/signup` | Create an account |
| `POST /auth/verify-email` | Verify the emailed code |
| `POST /auth/login` | Obtain authentication tokens |
| `POST /urls` | Create a shortened URL (authenticated) |
| `GET /urls` | List your links (authenticated) |
| `DELETE /urls/:id` | Delete an owned link |
| `GET /urls/:id/analytics` | Read click analytics (authenticated) |
| `GET /:short_code` | Redirect to the original URL |

After signing in, create a link with your access token:

```sh
curl -X POST http://localhost:8080/urls \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"long_url":"https://example.com","custom_alias":"example-link"}'
```

## Tests and limitations

Run `go test -race ./...`; the service regression tests need no external services. They cover database/cache redirects, click identity, expired/missing links, alias collisions and delete ownership/cache invalidation. CI runs the same command. These tests do not replace a full end-to-end authentication/database suite.

Click events use a bounded in-memory queue: a full queue or process restart can lose analytics. This is a learning project, not a guarantee of durable analytics. JWT v4 migration and database/broker major upgrades should be separately tested.
