-- +goose Up
CREATE TABLE keepalive_pings (
    id BIGSERIAL PRIMARY KEY,
    pinged_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE keepalive_pings;
