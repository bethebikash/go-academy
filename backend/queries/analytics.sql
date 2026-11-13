-- name: GetUserStreak :one
SELECT user_id, current_streak, longest_streak, last_activity_date, updated_at
FROM user_streaks
WHERE user_id = $1;

-- name: UpsertUserStreak :one
INSERT INTO user_streaks (
    user_id,
    current_streak,
    longest_streak,
    last_activity_date,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW()
)
ON CONFLICT (user_id) DO UPDATE SET
    current_streak = EXCLUDED.current_streak,
    longest_streak = EXCLUDED.longest_streak,
    last_activity_date = EXCLUDED.last_activity_date,
    updated_at = NOW()
RETURNING user_id, current_streak, longest_streak, last_activity_date, updated_at;

-- name: RecordUserActivity :one
INSERT INTO user_activities (
    user_id,
    activity_type,
    activity_date,
    metadata
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, user_id, activity_type, activity_date, metadata, created_at;

-- name: ListUserActivitiesRecent :many
SELECT id, user_id, activity_type, activity_date, metadata, created_at
FROM user_activities
WHERE user_id = $1 AND activity_date >= $2
ORDER BY activity_date DESC, created_at DESC;

-- name: CountUserActivitiesByDate :many
SELECT activity_date, COUNT(*)::int as count
FROM user_activities
WHERE user_id = $1 AND activity_date >= $2
GROUP BY activity_date
ORDER BY activity_date ASC;

-- name: AwardBadge :one
INSERT INTO user_badges (
    user_id,
    badge_key
) VALUES (
    $1, $2
)
ON CONFLICT (user_id, badge_key) DO NOTHING
RETURNING id, user_id, badge_key, awarded_at;

-- name: ListUserBadges :many
SELECT id, user_id, badge_key, awarded_at
FROM user_badges
WHERE user_id = $1
ORDER BY awarded_at ASC;

-- name: HasBadge :one
SELECT EXISTS(
    SELECT 1 FROM user_badges
    WHERE user_id = $1 AND badge_key = $2
);

-- name: CreateCertificate :one
INSERT INTO certificates (
    user_id,
    learning_path_id,
    certificate_code,
    recipient_name,
    learning_path_title
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, user_id, learning_path_id, certificate_code, recipient_name, learning_path_title, issued_at;

-- name: GetCertificateByCode :one
SELECT id, user_id, learning_path_id, certificate_code, recipient_name, learning_path_title, issued_at
FROM certificates
WHERE certificate_code = $1;

-- name: GetCertificateByUserAndPath :one
SELECT id, user_id, learning_path_id, certificate_code, recipient_name, learning_path_title, issued_at
FROM certificates
WHERE user_id = $1 AND learning_path_id = $2;

-- name: ListCertificatesByUser :many
SELECT id, user_id, learning_path_id, certificate_code, recipient_name, learning_path_title, issued_at
FROM certificates
WHERE user_id = $1
ORDER BY issued_at DESC;

-- name: GetUserStats :one
SELECT
    (SELECT COUNT(*)::int FROM lesson_progress lp WHERE lp.user_id = $1 AND lp.status = 'completed') as completed_lessons,
    (SELECT COUNT(DISTINCT cs.challenge_id)::int FROM challenge_submissions cs WHERE cs.user_id = $1 AND cs.status = 'passed') as solved_challenges,
    (SELECT COUNT(*)::int FROM notes n WHERE n.user_id = $1) as total_notes,
    (SELECT COUNT(*)::int FROM certificates c WHERE c.user_id = $1) as total_certificates;
