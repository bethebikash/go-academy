-- Bookmarks

-- name: CreateLessonBookmark :one
INSERT INTO bookmarks (user_id, lesson_id)
VALUES ($1, $2)
ON CONFLICT (user_id, lesson_id) DO NOTHING
RETURNING *;

-- name: GetLessonBookmark :one
SELECT * FROM bookmarks
WHERE user_id = $1 AND lesson_id = $2;

-- name: ListUserLessonBookmarks :many
SELECT 
    b.id,
    b.user_id,
    b.lesson_id,
    b.created_at,
    l.title AS lesson_title,
    l.slug AS lesson_slug,
    l.difficulty AS lesson_difficulty,
    l.estimated_minutes AS lesson_estimated_minutes,
    COALESCE(l.summary, '')::varchar AS lesson_summary,
    c.title AS course_title,
    c.slug AS course_slug,
    COALESCE(lp.status, 'not_started')::varchar AS progress_status
FROM bookmarks b
JOIN lessons l ON l.id = b.lesson_id
JOIN modules m ON m.id = l.module_id
JOIN courses c ON c.id = m.course_id
LEFT JOIN lesson_progress lp ON lp.lesson_id = l.id AND lp.user_id = b.user_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC;

-- name: DeleteBookmarkByID :exec
DELETE FROM bookmarks
WHERE id = $1 AND user_id = $2;

-- name: DeleteLessonBookmark :exec
DELETE FROM bookmarks
WHERE user_id = $1 AND lesson_id = $2;
