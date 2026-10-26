-- Use pgcrypto for UUID generation. It is widely available on managed PostgreSQL
-- providers and avoids depending on uuid-ossp installation in restricted environments.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('USER', 'ADMIN');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'USER';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'ADMIN';

CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ NULL,
    role user_role NOT NULL DEFAULT 'USER',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- most recent successful login metadata
    last_login_at TIMESTAMPTZ NULL,
    last_login_ip INET NULL,
    last_login_user_agent TEXT NULL,
    last_login_device TEXT NULL,

    -- optional profile fields
    profile_picture_url TEXT NULL,
    bio TEXT NULL,
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number TEXT NULL,

    metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_users PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users (id) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_email_active ON users (email) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users (last_login_at DESC);

-- Backward-compat cleanup: remove legacy token tables from older schema versions.
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS tokens;

CREATE TABLE IF NOT EXISTS sessions (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    access_token_hash TEXT NOT NULL,
    refresh_token_hash TEXT NOT NULL,
    access_expires_at TIMESTAMPTZ NOT NULL,
    refresh_expires_at TIMESTAMPTZ NOT NULL,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    ip_address INET NULL,
    user_agent TEXT NULL,
    device TEXT NULL,
    metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_sessions PRIMARY KEY (id),
    CONSTRAINT ux_sessions_access_hash UNIQUE (access_token_hash),
    CONSTRAINT ux_sessions_refresh_hash UNIQUE (refresh_token_hash)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id_active ON sessions (user_id) WHERE is_revoked = FALSE;
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires_at ON sessions (refresh_expires_at);