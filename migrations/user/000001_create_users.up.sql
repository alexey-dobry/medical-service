CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    first_name  TEXT        NOT NULL,
    middle_name TEXT        NOT NULL,
    last_name   TEXT        NOT NULL,
    phone       TEXT        NOT NULL,
    email       TEXT        NOT NULL UNIQUE,
    sex         TEXT        NOT NULL,
    birth_date  DATE        NOT NULL,
    photo_id    UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS doctors (
    user_id         UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    specialty       TEXT    NOT NULL,
    work_experience INTEGER NOT NULL CHECK (work_experience >= 0),
    description     TEXT    NOT NULL,
    services        TEXT[]  NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS photo_meta (
    id          UUID PRIMARY KEY,
    user_id     UUID   NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT   NOT NULL,
    mime_type   TEXT   NOT NULL,
    size        BIGINT NOT NULL,
    storage_key TEXT   NOT NULL UNIQUE
);

CREATE INDEX IF NOT EXISTS photo_meta_user_id_idx ON photo_meta (user_id);
