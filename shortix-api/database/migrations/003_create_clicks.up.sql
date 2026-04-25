CREATE TABLE IF NOT EXISTS clicks (
    id BIGSERIAL NOT NULL,
    url_id UUID NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET NULL,
    user_agent TEXT NULL,
    device TEXT NULL,
    referrer TEXT NULL,

    -- metadata
    metadata JSONB NULL,

    -- Constraints
    CONSTRAINT pk_clicks PRIMARY KEY (id),
    CONSTRAINT fk_clicks_url_id__urls_id
        FOREIGN KEY (url_id)
        REFERENCES urls (id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS ix_clicks_url_id ON clicks (url_id);
CREATE INDEX IF NOT EXISTS ix_clicks_clicked_at ON clicks (clicked_at);
