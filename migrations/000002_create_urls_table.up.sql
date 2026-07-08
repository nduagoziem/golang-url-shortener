CREATE TABLE urls (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shortened_url_code   TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE clicks (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url_id     UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT
);

CREATE INDEX idx_clicks_url_id   ON clicks(url_id);
