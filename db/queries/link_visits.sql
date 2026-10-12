-- name: CreateLinkVisit :one
INSERT INTO link_visits(link_id, ip, user_agent, referer, status)
VALUES (@link_id, @ip, @user_agent, @referer, @status)
RETURNING id, link_id, created_at, ip, user_agent, referer, status;

-- name: CountLinkVisits :one
SELECT COUNT(id) FROM link_visits;

-- name: GetLinkVisitsPage :many
-- One page in the requested order, as GetLinksPage. referer keeps visits
-- without one last in both directions.
SELECT
    id,
    link_id,
    created_at,
    ip,
    user_agent,
    referer,
    status
FROM link_visits
ORDER BY
    CASE WHEN @sort_field::text = 'link_id'
        AND @sort_asc::boolean
        THEN link_id END ASC,
    CASE WHEN @sort_field::text = 'link_id'
        AND NOT @sort_asc::boolean
        THEN link_id END DESC,
    CASE WHEN @sort_field::text = 'created_at'
        AND @sort_asc::boolean
        THEN created_at END ASC,
    CASE WHEN @sort_field::text = 'created_at'
        AND NOT @sort_asc::boolean
        THEN created_at END DESC,
    CASE WHEN @sort_field::text = 'ip'
        AND @sort_asc::boolean
        THEN ip END ASC,
    CASE WHEN @sort_field::text = 'ip'
        AND NOT @sort_asc::boolean
        THEN ip END DESC,
    CASE WHEN @sort_field::text = 'user_agent'
        AND @sort_asc::boolean
        THEN user_agent END ASC,
    CASE WHEN @sort_field::text = 'user_agent'
        AND NOT @sort_asc::boolean
        THEN user_agent END DESC,
    CASE WHEN @sort_field::text = 'status'
        AND @sort_asc::boolean
        THEN status END ASC,
    CASE WHEN @sort_field::text = 'status'
        AND NOT @sort_asc::boolean
        THEN status END DESC,
    CASE WHEN @sort_field::text = 'referer'
        AND @sort_asc::boolean
        THEN referer END ASC NULLS LAST,
    CASE WHEN @sort_field::text = 'referer'
        AND NOT @sort_asc::boolean
        THEN referer END DESC NULLS LAST,
    CASE WHEN @sort_field::text = 'id'
        AND NOT @sort_asc::boolean
        THEN id END DESC,
    id ASC
OFFSET sqlc.arg('offset')::bigint
LIMIT sqlc.arg('limit')::bigint;
