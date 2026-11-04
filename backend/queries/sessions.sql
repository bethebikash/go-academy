-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (
    user_id,
    token_hash,
    token_family_id,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING id, user_id, token_hash, token_family_id, expires_at, created_at;

-- name: GetRefreshSessionByTokenHash :one
SELECT id, user_id, token_hash, token_family_id, expires_at, revoked_at, created_at, last_used_at
FROM refresh_sessions
WHERE token_hash = $1;

-- name: RevokeRefreshSession :exec
UPDATE refresh_sessions
SET revoked_at = NOW(), last_used_at = NOW()
WHERE id = $1;

-- name: RevokeRefreshFamily :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE token_family_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredSessions :exec
DELETE FROM refresh_sessions
WHERE expires_at < NOW() OR revoked_at < NOW() - INTERVAL '30 days';
