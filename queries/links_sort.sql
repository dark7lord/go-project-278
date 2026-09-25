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