-- name: CreateLink :one
INSERT INTO links (original_url, short_name)
VALUES (@original_url, @short_name)
RETURNING id, original_url, short_name, created_at;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE id = @id;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE short_name = @short_name;

-- name: CountLinks :one
SELECT COUNT(id) FROM links;

-- name: UpdateLink :one
UPDATE links
SET
    original_url = @original_url,
    short_name = @short_name
WHERE id = @id
RETURNING id, original_url, short_name, created_at;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = @id
RETURNING id, original_url, short_name, created_at;

-- name: GetLinksPage :many
-- One page in the requested order. Each CASE is live for one field and
-- direction and NULL otherwise; id comes last, so ties and no sort go by id.
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY
    CASE WHEN sqlc.arg(sort_field)::text = 'short_name'
        AND sqlc.arg(sort_asc)::boolean
        THEN short_name END ASC,
    CASE WHEN sqlc.arg(sort_field)::text = 'short_name'
        AND NOT sqlc.arg(sort_asc)::boolean
        THEN short_name END DESC,
    CASE WHEN sqlc.arg(sort_field)::text = 'original_url'
        AND sqlc.arg(sort_asc)::boolean
        THEN original_url END ASC,
    CASE WHEN sqlc.arg(sort_field)::text = 'original_url'
        AND NOT sqlc.arg(sort_asc)::boolean
        THEN original_url END DESC,
    CASE WHEN sqlc.arg(sort_field)::text = 'id'
        AND NOT sqlc.arg(sort_asc)::boolean
        THEN id END DESC,
    id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;
