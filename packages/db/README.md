# postgres — queuepost-engine

Postgres schema for `outpost-engine` (modular monolith).

## Schema
- `migrations/001_initial_schema.sql` — `users`, `social_accounts`, `posts` (Phase 1, matches `apps/server/internal/models/models.go:1`)
- `migrations/002_seed_demo.sql` — demo user + linkedin account + 2 posts

Details:
- `media_urls` is `JSONB` (`[]string` in Go, `[]` default) — aligned with `Post.MediaUrls` and `handlers.go` insert.
- `status` `CHECK (PENDING|SCHEDULED|PUBLISHED|FAILED)` + index `idx_posts_status_scheduled` for worker polling.
- `social_accounts.access_token` plain `TEXT` — encrypt at app layer (AGENTS.md workflow #1).
- `updated_at` auto-trigger.

## How to run (Postgres already running, DB from `apps/server/.env`)

```bash
# 1. create DB if not exists (uses DB_* from apps/server/.env)
createdb queuepost_db  # or: psql -h localhost -U postgres -c "CREATE DATABASE queuepost_db;"

# 2. run migrations (psql)
psql "host=localhost port=5432 user=postgres password=secret dbname=queuepost_db sslmode=disable" \
  -f packages/db/migrations/001_initial_schema.sql

# optional seed
psql "host=localhost port=5432 user=postgres password=secret dbname=queuepost_db sslmode=disable" \
  -f packages/db/migrations/002_seed_demo.sql

# 3. verify
psql "host=localhost port=5432 user=postgres password=secret dbname=queuepost_db sslmode=disable" \
  -c "\d posts" -c "SELECT * FROM posts LIMIT 5;"
```

Or with `DATABASE_URL`:
```bash
psql $DATABASE_URL -f packages/db/migrations/001_initial_schema.sql
```

## BE wiring
- DSN built in `apps/server/internal/database/db.go:13` from `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSLMODE`
- `models.Post.MediaUrls` expects `JSONB` — `handlers.go` should `json.Marshal` on insert (currently inserts slice directly; will be fixed to `pq`/`jsonb`).
- Worker `internal/worker/processor.go:43` queries `posts` + `social_accounts` with same schema.
