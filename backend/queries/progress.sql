-- Lesson Progress

-- name: GetLessonProgress :one
SELECT * FROM lesson_progress
WHERE user_id = $1 AND lesson_id = $2;

-- name: UpsertLessonProgress :one
INSERT INTO lesson_progress (
    user_id, lesson_id, status, started_at, completed_at, last_accessed_at, updated_at
) VALUES (
    $1, $2, $3, NOW(), $4, NOW(), NOW()
)
ON CONFLICT (user_id, lesson_id) DO UPDATE
SET
    status = CASE 
        WHEN lesson_progress.status = 'completed' THEN 'completed' 
        ELSE EXCLUDED.status 
    END,
    completed_at = CASE 
        WHEN lesson_progress.completed_at IS NOT NULL THEN lesson_progress.completed_at 
        ELSE EXCLUDED.completed_at 
    END,
    last_accessed_at = NOW(),
    updated_at = NOW()
RETURNING *;

-- name: ListUserProgressForCourse :many
SELECT lp.* FROM lesson_progress lp
JOIN lessons l ON l.id = lp.lesson_id
JOIN modules m ON m.id = l.module_id
WHERE lp.user_id = $1 AND m.course_id = $2;

-- name: GetCourseProgressStats :one
SELECT 
    COUNT(l.id)::int AS total_lessons,
    COUNT(lp.id) FILTER (WHERE lp.status = 'completed')::int AS completed_lessons
FROM lessons l
JOIN modules m ON m.id = l.module_id
LEFT JOIN lesson_progress lp ON lp.lesson_id = l.id AND lp.user_id = $1
WHERE m.course_id = $2 AND l.status = 'published';

-- name: GetLearningPathProgressStats :one
SELECT 
    COUNT(l.id)::int AS total_lessons,
    COUNT(lp.id) FILTER (WHERE lp.status = 'completed')::int AS completed_lessons
FROM lessons l
JOIN modules m ON m.id = l.module_id
JOIN courses c ON c.id = m.course_id
LEFT JOIN lesson_progress lp ON lp.lesson_id = l.id AND lp.user_id = $1
WHERE c.learning_path_id = $2 AND c.status = 'published' AND l.status = 'published';
