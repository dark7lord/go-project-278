-- +goose Up
ALTER TABLE links
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL;


-- +goose Down
ALTER TABLE links
    DROP COLUMN created_at;
