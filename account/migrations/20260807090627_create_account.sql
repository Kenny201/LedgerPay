-- +goose Up
CREATE TABLE accounts (
                          id            BIGSERIAL PRIMARY KEY,
                          user_id       BIGINT      NOT NULL REFERENCES users (id),
                          currency      CHAR(3)     NOT NULL DEFAULT 'RUB',
                          balance_cents BIGINT      NOT NULL DEFAULT 0,
                          updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
                          CONSTRAINT accounts_user_id_key UNIQUE (user_id),
                          CONSTRAINT accounts_balance_nonnegative CHECK (balance_cents >= 0)
);

-- +goose Down
DROP TABLE IF EXISTS accounts;