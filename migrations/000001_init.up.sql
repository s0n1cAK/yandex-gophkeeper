CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    username      CITEXT NOT NULL UNIQUE,
    password_hash TEXT   NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS secrets (
    id         BIGSERIAL PRIMARY KEY,
    owner_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    type       TEXT NOT NULL CHECK (type IN ('password', 'text', 'bank_card', 'binary')),
    comment    TEXT NOT NULL DEFAULT '',

    key_id     TEXT  NOT NULL,
    nonce      BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_secrets_owner_updated
    ON secrets(owner_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_secrets_owner_type
    ON secrets(owner_id, type)
    WHERE deleted_at IS NULL;
