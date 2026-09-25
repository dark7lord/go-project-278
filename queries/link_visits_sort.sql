-- name: GetLinkVisitsRangeCreatedAtDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY created_at DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeCreatedAtAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY created_at ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeLinkIdDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY link_id DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeLinkIdAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY link_id ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeIpDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY ip DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeIpAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY ip ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeUserAgentDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY user_agent DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeUserAgentAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY user_agent ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeStatusDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY status DESC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeStatusAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY status ASC, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeRefererAsc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY referer ASC NULLS LAST, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeRefererDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY referer DESC NULLS LAST, id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;

-- name: GetLinkVisitsRangeIdDesc :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY id DESC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;