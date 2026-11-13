-- Notes

-- name: CreateNote :one
INSERT INTO notes (
    user_id, lesson_id, title, content, pinned, version
) VALUES (
    $1, $2, $3, $4, $5, 1
) RETURNING *;

-- name: GetNoteByID :one
SELECT * FROM notes
WHERE id = $1 AND user_id = $2;

-- name: ListNotesByUserID :many
SELECT * FROM notes
WHERE user_id = $1
ORDER BY pinned DESC, updated_at DESC;

-- name: ListNotesByLessonID :many
SELECT * FROM notes
WHERE user_id = $1 AND lesson_id = $2
ORDER BY pinned DESC, updated_at DESC;

-- name: SearchNotes :many
SELECT * FROM notes
WHERE user_id = $1
  AND to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $2)
ORDER BY pinned DESC, updated_at DESC;

-- name: UpdateNote :one
UPDATE notes
SET
    title = $4,
    content = $5,
    pinned = $6,
    version = version + 1,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND version = $3
RETURNING *;

-- name: DeleteNote :exec
DELETE FROM notes
WHERE id = $1 AND user_id = $2;

-- Tags

-- name: UpsertTag :one
INSERT INTO tags (user_id, name)
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE
SET name = EXCLUDED.name
RETURNING *;

-- name: ListUserTags :many
SELECT t.id, t.name, t.created_at, COUNT(nt.note_id)::int AS note_count
FROM tags t
LEFT JOIN note_tags nt ON nt.tag_id = t.id
WHERE t.user_id = $1
GROUP BY t.id, t.name, t.created_at
ORDER BY t.name ASC;

-- name: AddNoteTag :exec
INSERT INTO note_tags (note_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ClearNoteTags :exec
DELETE FROM note_tags
WHERE note_id = $1;

-- name: GetNoteTags :many
SELECT t.* FROM tags t
JOIN note_tags nt ON nt.tag_id = t.id
WHERE nt.note_id = $1
ORDER BY t.name ASC;

-- name: ListNotesByTag :many
SELECT n.* FROM notes n
JOIN note_tags nt ON nt.note_id = n.id
JOIN tags t ON t.id = nt.tag_id
WHERE n.user_id = $1 AND t.name = $2
ORDER BY n.pinned DESC, n.updated_at DESC;
