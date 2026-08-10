-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS link_visits (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    link_id         BIGINT REFERENCES links(id) ON DELETE CASCADE NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    ip              VARCHAR(45) NOT NULL DEFAULT '',
    user_agent      TEXT NOT NULL DEFAULT '',
    referer         TEXT,
    status          INT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link_visits;
-- +goose StatementEnd
