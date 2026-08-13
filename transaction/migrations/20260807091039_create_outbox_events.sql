-- +goose Up
CREATE TABLE outbox_events (
                               id             BIGSERIAL PRIMARY KEY,
                               aggregate_type TEXT        NOT NULL,
                               aggregate_id   BIGINT      NOT NULL,
                               event_type     TEXT        NOT NULL,
                               payload        JSONB       NOT NULL,
                               created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
                               published_at   TIMESTAMPTZ NULL
);

-- +goose Down
DROP TABLE IF EXISTS outbox_events;