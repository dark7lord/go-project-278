-- name: CreateLink :one
INSERT INTO links (original_url, short_name, short_url)
VALUES (@original_url, @short_name, @short_url)
RETURNING id, original_url, short_name, short_url;

-- name: GetLink :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE id = @id;

-- name: GetLinks :many
SELECT id, original_url, short_name, short_url
FROM links;

-- name: GetLinksRange :many
SELECT id, original_url, short_name, short_url
FROM links
ORDER BY id
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: CountLinks :one
SELECT COUNT(id) FROM links;

-- name: GetShortNames :many
SELECT short_name FROM links;

-- name: UpdateLink :one
UPDATE links
SET
    original_url = @original_url,
    short_name = @short_name,
    short_url = @short_url
WHERE id = @id
RETURNING id, original_url, short_name, short_url;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = @id
RETURNING id, original_url, short_name, short_url;