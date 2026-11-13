-- Challenges

-- name: ListPublishedChallenges :many
SELECT 
    id, course_id, lesson_id, title, slug, description, difficulty, 
    estimated_minutes, starter_code, status, sort_order, created_at, updated_at, published_at
FROM challenges
WHERE status = 'published'
  AND ($1::varchar = '' OR difficulty = $1)
ORDER BY sort_order ASC, created_at ASC;

-- name: ListPublishedChallengesByCourseID :many
SELECT 
    id, course_id, lesson_id, title, slug, description, difficulty, 
    estimated_minutes, starter_code, status, sort_order, created_at, updated_at, published_at
FROM challenges
WHERE status = 'published' AND course_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: GetPublishedChallengeBySlug :one
SELECT 
    id, course_id, lesson_id, title, slug, description, difficulty, 
    estimated_minutes, starter_code, status, sort_order, created_at, updated_at, published_at
FROM challenges
WHERE slug = $1 AND status = 'published';

-- name: ListAllChallenges :many
SELECT * FROM challenges
ORDER BY sort_order ASC, created_at ASC;

-- name: GetChallengeByID :one
SELECT * FROM challenges
WHERE id = $1;

-- name: GetChallengeBySlug :one
SELECT * FROM challenges
WHERE slug = $1;

-- name: CreateChallenge :one
INSERT INTO challenges (
    course_id, lesson_id, title, slug, description, difficulty, estimated_minutes,
    starter_code, solution_code, status, sort_order, published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateChallenge :one
UPDATE challenges
SET
    course_id = $2,
    lesson_id = $3,
    title = $4,
    slug = $5,
    description = $6,
    difficulty = $7,
    estimated_minutes = $8,
    starter_code = $9,
    solution_code = $10,
    sort_order = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateChallengeStatus :one
UPDATE challenges
SET
    status = $2,
    published_at = COALESCE($3, published_at),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteChallenge :exec
DELETE FROM challenges
WHERE id = $1;

-- Test Cases

-- name: ListSampleTestCasesByChallengeID :many
SELECT id, challenge_id, title, input, expected_output, hidden, sort_order, created_at
FROM challenge_test_cases
WHERE challenge_id = $1 AND hidden = false
ORDER BY sort_order ASC, created_at ASC;

-- name: ListAllTestCasesByChallengeID :many
SELECT * FROM challenge_test_cases
WHERE challenge_id = $1
ORDER BY sort_order ASC, created_at ASC;

-- name: CreateTestCase :one
INSERT INTO challenge_test_cases (
    challenge_id, title, input, expected_output, hidden, sort_order
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ClearTestCases :exec
DELETE FROM challenge_test_cases
WHERE challenge_id = $1;
