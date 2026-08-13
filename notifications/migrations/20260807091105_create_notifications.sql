-- +goose Up
CREATE TABLE notifications (
                               id         BIGSERIAL PRIMARY KEY,
                               user_id    BIGINT      NOT NULL,
                               type       TEXT        NOT NULL,
                               payload    JSONB       NOT NULL,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS notifications;