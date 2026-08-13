-- +goose Up
CREATE TABLE users (
                       id            BIGSERIAL PRIMARY KEY,
                       login         TEXT        NOT NULL,
                       email         TEXT        NOT NULL,
                       password_hash TEXT        NOT NULL,
                       created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
                       CONSTRAINT users_login_key UNIQUE (login),
                       CONSTRAINT users_email_key UNIQUE (email)
);

-- +goose Down
DROP TABLE IF EXISTS users;