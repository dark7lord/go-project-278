-- name: CreateLinkVisit :one
INSERT INTO link_visits(link_id, ip, user_agent, referer, status)
VALUES (@link_id, @ip, @user_agent, @referer, @status)
RETURNING id, link_id, created_at, ip, user_agent, referer, status;

-- name: CountLinkVisits :one
SELECT COUNT(id) FROM link_visits;

-- name: GetLinkVisits :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY id;

-- name: GetLinkVisitsRange :many
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY id
OFFSET sqlc.arg('offset')
LIMIT sqlc.arg('limit');
