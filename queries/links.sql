-- name: CreateLink :one
INSERT INTO links (original_url, short_name)
VALUES (@original_url, @short_name)
RETURNING id, original_url, short_name;

-- name: GetLinkByID :one
SELECT id, original_url, short_name
FROM links
WHERE id = @id;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name
FROM links
WHERE short_name = @short_name;

-- name: GetLinks :many
SELECT id, original_url, short_name
FROM links;

-- name: GetLinksRange :many
SELECT id, original_url, short_name
FROM links
ORDER BY id
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: CountLinks :one
SELECT COUNT(id) FROM links;

-- name: UpdateLink :one
UPDATE links
SET
    original_url = @original_url,
    short_name = @short_name
WHERE id = @id
RETURNING id, original_url, short_name;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = @id
RETURNING id, original_url, short_name;