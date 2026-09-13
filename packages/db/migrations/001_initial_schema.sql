-- 001_initial_schema.sql
-- Postgres schema for queuepost-engine (Phase 1: users, social_accounts, posts)
-- Compatible with apps/server/internal/models/models.go:1 and lib/pq + database/sql

-- Enable UUID if needed later (optional)
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- users: auth owner for posts
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- social_accounts: OAuth 2.0 tokens (encrypted at app layer, see AGENTS.md workflow #1)
CREATE TABLE IF NOT EXISTS social_accounts (
    id                SERIAL PRIMARY KEY,
    user_id           INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform          VARCHAR(50) NOT NULL, -- linkedin | instagram | etc
    platform_user_id  VARCHAR(255) NOT NULL,
    access_token      TEXT NOT NULL, -- store encrypted
    refresh_token     TEXT,
    token_expires_at  TIMESTAMPTZ NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, platform, platform_user_id)
);
CREATE INDEX IF NOT EXISTS idx_social_accounts_user_id ON social_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_social_accounts_platform ON social_accounts(platform);

-- posts: scheduled content, processed by Redis + Asynq worker
CREATE TABLE IF NOT EXISTS posts (
    id                  SERIAL PRIMARY KEY,
    user_id             INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    social_account_id   INT NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    content             TEXT NOT NULL,
    media_urls          JSONB NOT NULL DEFAULT '[]'::jsonb, -- []string from models.Post.MediaUrls
    status              VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING','SCHEDULED','PUBLISHED','FAILED')),
    scheduled_at        TIMESTAMPTZ NOT NULL,
    published_at        TIMESTAMPTZ,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_social_account_id ON posts(social_account_id);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_scheduled_at ON posts(scheduled_at);
-- Worker query optimization: fetch due posts ordered by schedule
CREATE INDEX IF NOT EXISTS idx_posts_status_scheduled ON posts(status, scheduled_at) WHERE status IN ('PENDING','SCHEDULED');

-- updated_at trigger helper
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_updated ON users;
CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_social_accounts_updated ON social_accounts;
CREATE TRIGGER trg_social_accounts_updated BEFORE UPDATE ON social_accounts FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_posts_updated ON posts;
CREATE TRIGGER trg_posts_updated BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION set_updated_at();
