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

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = @id;