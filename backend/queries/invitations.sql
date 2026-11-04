-- name: CreateInvitation :one
INSERT INTO invitations (
    email,
    token_hash,
    invited_by,
    expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING id, email, token_hash, invited_by, expires_at, created_at;

-- name: GetInvitationByTokenHash :one
SELECT id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at
FROM invitations
WHERE token_hash = $1;

-- name: GetInvitationByTokenHashForUpdate :one
SELECT id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at
FROM invitations
WHERE token_hash = $1
FOR UPDATE;

-- name: MarkInvitationAccepted :exec
UPDATE invitations
SET accepted_at = NOW()
WHERE id = $1;

-- name: RevokeInvitation :one
UPDATE invitations
SET revoked_at = NOW()
WHERE id = $1 AND accepted_at IS NULL AND revoked_at IS NULL
RETURNING id, email, revoked_at;

-- name: ListInvitations :many
SELECT i.id, i.email, i.token_hash, i.invited_by, i.expires_at, i.accepted_at, i.revoked_at, i.created_at,
       u.username AS invited_by_username
FROM invitations i
JOIN users u ON i.invited_by = u.id
ORDER BY i.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountInvitations :one
SELECT COUNT(*) FROM invitations;

-- name: GetInvitationByID :one
SELECT id, email, token_hash, invited_by, expires_at, accepted_at, revoked_at, created_at
FROM invitations
WHERE id = $1;
