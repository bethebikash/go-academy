-- 000006_add_gamification_and_analytics.up.sql
-- Create tables for streaks, daily activities, certificates, and badges

CREATE TABLE IF NOT EXISTS user_streaks (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT NOT NULL DEFAULT 0,
    longest_streak INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(64) NOT NULL,
    activity_date DATE NOT NULL DEFAULT CURRENT_DATE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_activities_user_date 
    ON user_activities(user_id, activity_date DESC);

CREATE TABLE IF NOT EXISTS certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    learning_path_id UUID NOT NULL REFERENCES learning_paths(id) ON DELETE CASCADE,
    certificate_code VARCHAR(64) UNIQUE NOT NULL,
    recipient_name VARCHAR(255) NOT NULL,
    learning_path_title VARCHAR(255) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_learning_path_certificate UNIQUE(user_id, learning_path_id)
);

CREATE TABLE IF NOT EXISTS user_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_key VARCHAR(64) NOT NULL,
    awarded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_badge UNIQUE(user_id, badge_key)
);

CREATE INDEX IF NOT EXISTS idx_user_badges_user 
    ON user_badges(user_id);
