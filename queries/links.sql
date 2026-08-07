-- name: CreateLink :one
INSERT INTO links (original_url, short_name, short_url)
VALUES (@original_url, @short_name, @short_url)
RETURNING id, original_url, short_name, short_url;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE id = @id;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, short_url
FROM links
WHERE short_name = @short_name;

-- name: GetLinks :many
SELECT id, original_url, short_name, short_url
FROM links;

-- name: GetLinksRange :many
SELECT id, original_url, short_name, short_url
FROM links
ORDER BY id
OFFSET sqlc.arg('offset')
LIMIT sqlc.arg('limit');

-- name: CountLinks :one
SELECT COUNT(id) FROM links;

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