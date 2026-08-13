-- +goose Up
CREATE TABLE ledger_entries (
                                id           BIGSERIAL PRIMARY KEY,
                                account_id   BIGINT      NOT NULL,
                                transfer_id  BIGINT      NULL REFERENCES transfers (id),
                                kind         TEXT        NOT NULL,
                                amount_cents BIGINT      NOT NULL,
                                created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
                                CONSTRAINT ledger_entries_amount_positive CHECK (amount_cents > 0),
                                CONSTRAINT ledger_entries_kind_check CHECK (kind IN ('topup', 'debit', 'credit')
                                    )
);

-- +goose Down
DROP TABLE IF EXISTS ledger_entries;