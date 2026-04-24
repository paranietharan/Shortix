CREATE TABLE IF NOT EXISTS urls (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NULL,
    long_url TEXT NOT NULL,
    short_code VARCHAR(64) NOT NULL,
    custom_alias VARCHAR(64) NULL,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- metadata
    metadata JSONB NULL,

    -- Constraints
    CONSTRAINT pk_urls PRIMARY KEY (id),
    CONSTRAINT fk_urls_user_id__users_id
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE SET NULL
);

-- short_code must be unique; this unique index also serves as the lookup index.
CREATE UNIQUE INDEX IF NOT EXISTS ux_urls_short_code ON urls (short_code);

-- Enforce uniqueness only when custom_alias is provided.
CREATE UNIQUE INDEX IF NOT EXISTS ux_urls_custom_alias_not_null
    ON urls (custom_alias)
    WHERE custom_alias IS NOT NULL;
