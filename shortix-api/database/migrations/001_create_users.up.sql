-- Use pgcrypto for UUID generation. It is widely available on managed PostgreSQL
-- providers and avoids depending on uuid-ossp installation in restricted environments.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE IF NOT EXISTS users (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- optional fields for future features
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ NULL,
    profile_picture_url TEXT NULL,
    bio TEXT NULL,
    is_premium BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number TEXT NULL,
    role user_role NOT NULL DEFAULT 'user',

    -- metadata
    metadata JSONB NULL,

    -- Constraints
    CONSTRAINT pk_users PRIMARY KEY (id)
);

-- Unique index enforces one account per email and supports login lookups.
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users (email);