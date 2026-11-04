-- name: GetRoleByName :one
SELECT id, name, description, created_at
FROM roles
WHERE name = $1;

-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: GetUserRoles :many
SELECT r.id, r.name, r.description
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;
