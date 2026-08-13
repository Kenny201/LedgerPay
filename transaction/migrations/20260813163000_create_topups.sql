-- +goose Up
CREATE TABLE topups (
    id              BIGSERIAL PRIMARY KEY,
    account_id      BIGINT      NOT NULL,
    amount_cents    BIGINT      NOT NULL,
    idempotency_key TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT topups_amount_positive CHECK (amount_cents > 0),
    CONSTRAINT topups_account_idempotency_key_unique UNIQUE (account_id, idempotency_key)
);

-- +goose Down
DROP TABLE IF EXISTS topups;
