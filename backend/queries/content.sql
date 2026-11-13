-- Learning Paths

-- name: ListPublishedLearningPaths :many
SELECT * FROM learning_paths
WHERE status = 'published'
ORDER BY sort_order ASC, created_at ASC;

-- name: ListAllLearningPaths :many
SELECT * FROM learning_paths
ORDER BY sort_order ASC, created_at ASC;

-- name: GetLearningPathByID :one
SELECT * FROM learning_paths
WHERE id = $1;

-- name: GetLearningPathBySlug :one
SELECT * FROM learning_paths
WHERE slug = $1;

-- name: CreateLearningPath :one
INSERT INTO learning_paths (
    title, slug, description, status, sort_order, created_by, published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateLearningPath :one
UPDATE learning_paths
SET 
    title = $2,
    slug = $3,
    description = $4,
    sort_order = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLearningPathStatus :one
UPDATE learning_paths
SET 
    status = $2,
    published_at = COALESCE($3, published_at),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLearningPath :exec
DELETE FROM learning_paths
WHERE id = $1;

-- Courses

-- name: ListPublishedCoursesByPathID :many
SELECT * FROM courses
WHERE learning_path_id = $1 AND status = 'published'
ORDER BY sort_order ASC, created_at ASC;

-- name: ListAllCoursesByPathID :many
SELECT * FROM courses
WHERE learning_path_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: GetCourseByID :one
SELECT * FROM courses
WHERE id = $1;

-- name: GetCourseBySlug :one
SELECT * FROM courses
WHERE slug = $1;

-- name: CreateCourse :one
INSERT INTO courses (
    learning_path_id, title, slug, description, status, sort_order, published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateCourse :one
UPDATE courses
SET 
    title = $2,
    slug = $3,
    description = $4,
    sort_order = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCourseStatus :one
UPDATE courses
SET 
    status = $2,
    published_at = COALESCE($3, published_at),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCourse :exec
DELETE FROM courses
WHERE id = $1;

-- Modules

-- name: ListModulesByCourseID :many
SELECT * FROM modules
WHERE course_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: GetModuleByID :one
SELECT * FROM modules
WHERE id = $1;

-- name: GetModuleBySlug :one
SELECT * FROM modules
WHERE course_id = $1 AND slug = $2;

-- name: CreateModule :one
INSERT INTO modules (
    course_id, title, slug, description, sort_order
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateModule :one
UPDATE modules
SET 
    title = $2,
    slug = $3,
    description = $4,
    sort_order = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteModule :exec
DELETE FROM modules
WHERE id = $1;

-- Lessons

-- name: ListPublishedLessonsByModuleID :many
SELECT * FROM lessons
WHERE module_id = $1 AND status = 'published'
ORDER BY sort_order ASC, created_at ASC;

-- name: ListAllLessonsByModuleID :many
SELECT * FROM lessons
WHERE module_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: GetLessonByID :one
SELECT * FROM lessons
WHERE id = $1;

-- name: GetLessonBySlug :one
SELECT * FROM lessons
WHERE slug = $1;

-- name: CreateLesson :one
INSERT INTO lessons (
    module_id, title, slug, summary, content, difficulty, estimated_minutes, status, sort_order, published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateLesson :one
UPDATE lessons
SET 
    title = $2,
    slug = $3,
    summary = $4,
    content = $5,
    difficulty = $6,
    estimated_minutes = $7,
    sort_order = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLessonStatus :one
UPDATE lessons
SET 
    status = $2,
    published_at = COALESCE($3, published_at),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteLesson :exec
DELETE FROM lessons
WHERE id = $1;

-- Prerequisites

-- name: AddLessonPrerequisite :exec
INSERT INTO lesson_prerequisites (lesson_id, prerequisite_lesson_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveLessonPrerequisite :exec
DELETE FROM lesson_prerequisites
WHERE lesson_id = $1 AND prerequisite_lesson_id = $2;

-- name: ClearLessonPrerequisites :exec
DELETE FROM lesson_prerequisites
WHERE lesson_id = $1;

-- name: GetLessonPrerequisites :many
SELECT l.* FROM lessons l
JOIN lesson_prerequisites lp ON lp.prerequisite_lesson_id = l.id
WHERE lp.lesson_id = $1
ORDER BY l.sort_order ASC;

-- name: GetUncompletedPrerequisitesForUser :many
SELECT l.* FROM lessons l
JOIN lesson_prerequisites lp ON lp.prerequisite_lesson_id = l.id
LEFT JOIN lesson_progress prog ON prog.lesson_id = l.id AND prog.user_id = $2 AND prog.status = 'completed'
WHERE lp.lesson_id = $1 AND prog.id IS NULL;
