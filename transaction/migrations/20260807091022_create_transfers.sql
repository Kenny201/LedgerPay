-- +goose Up
CREATE TABLE transfers (
                           id               BIGSERIAL PRIMARY KEY,
                           from_account_id  BIGINT      NOT NULL,
                           to_account_id    BIGINT      NOT NULL,
                           amount_cents     BIGINT      NOT NULL,
                           idempotency_key  TEXT        NOT NULL,
                           status           TEXT        NOT NULL,
                           created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
                           CONSTRAINT transfers_amount_positive CHECK (amount_cents > 0),
                           CONSTRAINT transfers_status_check CHECK (status IN ('pending', 'completed', 'failed')),
                           CONSTRAINT transfers_from_account_idempotency_key_unique UNIQUE (from_account_id, idempotency_key)
);

-- +goose Down
DROP TABLE IF EXISTS transfers;