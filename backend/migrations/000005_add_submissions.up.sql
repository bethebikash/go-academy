-- 000005_add_submissions.up.sql
-- Create submissions and execution results tables

CREATE TABLE IF NOT EXISTS challenge_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    total_test_cases INT NOT NULL DEFAULT 0,
    passed_test_cases INT NOT NULL DEFAULT 0,
    execution_time_ms INT NOT NULL DEFAULT 0,
    error_output TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS challenge_submission_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES challenge_submissions(id) ON DELETE CASCADE,
    test_case_id UUID REFERENCES challenge_test_cases(id) ON DELETE SET NULL,
    test_case_title VARCHAR(255) NOT NULL,
    passed BOOLEAN NOT NULL DEFAULT FALSE,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    input TEXT NOT NULL DEFAULT '',
    expected_output TEXT NOT NULL DEFAULT '',
    actual_output TEXT NOT NULL DEFAULT '',
    execution_time_ms INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_challenge_submissions_user_challenge 
    ON challenge_submissions(user_id, challenge_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_challenge_submissions_challenge_status 
    ON challenge_submissions(challenge_id, status);

CREATE INDEX IF NOT EXISTS idx_submission_results_submission_id 
    ON challenge_submission_results(submission_id);
