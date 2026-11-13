-- name: CreateSubmission :one
INSERT INTO challenge_submissions (
    user_id,
    challenge_id,
    code,
    status,
    total_test_cases,
    passed_test_cases,
    execution_time_ms,
    error_output
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateSubmission :one
UPDATE challenge_submissions
SET
    status = $2,
    total_test_cases = $3,
    passed_test_cases = $4,
    execution_time_ms = $5,
    error_output = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSubmissionByID :one
SELECT * FROM challenge_submissions
WHERE id = $1;

-- name: ListSubmissionsByChallengeAndUser :many
SELECT * FROM challenge_submissions
WHERE user_id = $1 AND challenge_id = $2
ORDER BY created_at DESC
LIMIT 20;

-- name: CreateSubmissionResult :one
INSERT INTO challenge_submission_results (
    submission_id,
    test_case_id,
    test_case_title,
    passed,
    hidden,
    input,
    expected_output,
    actual_output,
    execution_time_ms,
    error_message
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListSubmissionResults :many
SELECT * FROM challenge_submission_results
WHERE submission_id = $1
ORDER BY created_at ASC;
