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

-- name: GetLinksRangeIdAsc :many
SELECT id, original_url, short_name
FROM links
ORDER BY id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinksRangeShortNameAsc :many
SELECT id, original_url, short_name
FROM links
ORDER BY short_name ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinksRangeShortNameDesc :many
SELECT id, original_url, short_name
FROM links
ORDER BY short_name DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinksRangeOriginalURLAsc :many
SELECT id, original_url, short_name
FROM links
ORDER BY original_url ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinksRangeOriginalURLDesc :many
SELECT id, original_url, short_name
FROM links
ORDER BY original_url DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinksRangeIdDesc :many
SELECT id, original_url, short_name
FROM links
ORDER BY id DESC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;
