-- name: GetUserByID :one
SELECT id, email, username, password_hash, status, timezone, last_login_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, username, password_hash, status, timezone, last_login_at, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1);

-- name: GetUserByUsername :one
SELECT id, email, username, password_hash, status, timezone, last_login_at, created_at, updated_at
FROM users
WHERE LOWER(username) = LOWER($1);

-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    password_hash,
    status,
    timezone
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, email, username, status, timezone, created_at, updated_at;

-- name: UpdateUserStatus :one
UPDATE users
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, email, username, status, updated_at;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: ListUsers :many
SELECT u.id, u.email, u.username, u.status, u.timezone, u.last_login_at, u.created_at, u.updated_at,
       COALESCE(r.name, 'learner') AS role
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CountAdmins :one
SELECT COUNT(*)
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE r.name = 'admin';
