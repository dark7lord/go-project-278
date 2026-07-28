-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS links (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    original_url    VARCHAR(2048) NOT NULL,
    short_name      VARCHAR(50) NOT NULL UNIQUE,
    short_url       VARCHAR(128) NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd