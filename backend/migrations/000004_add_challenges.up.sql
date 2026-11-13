-- 000004_add_challenges.up.sql

CREATE TABLE IF NOT EXISTS challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    difficulty VARCHAR(32) NOT NULL DEFAULT 'beginner', -- beginner, intermediate, advanced
    estimated_minutes INT NOT NULL DEFAULT 15,
    starter_code TEXT NOT NULL,
    solution_code TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft',       -- draft, published, archived
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_challenges_status_difficulty ON challenges (status, difficulty);
CREATE INDEX IF NOT EXISTS idx_challenges_course ON challenges (course_id);
CREATE INDEX IF NOT EXISTS idx_challenges_lesson ON challenges (lesson_id);

CREATE TABLE IF NOT EXISTS challenge_test_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    input TEXT NOT NULL,
    expected_output TEXT NOT NULL,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_test_cases_challenge ON challenge_test_cases (challenge_id, sort_order);

-- Foreign keys and constraints for notes and bookmarks
ALTER TABLE notes 
    ADD CONSTRAINT fk_notes_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE SET NULL;

ALTER TABLE bookmarks 
    ADD CONSTRAINT fk_bookmarks_challenge FOREIGN KEY (challenge_id) REFERENCES challenges(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_challenge_bookmark ON bookmarks (user_id, challenge_id) WHERE challenge_id IS NOT NULL;
