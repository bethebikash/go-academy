-- 000001_initial_schema.up.sql

-- Enable pgcrypto for UUID generation if needed (gen_random_uuid is built-in in PG 13+)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(username));
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(32) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed system roles safely
INSERT INTO roles (name, description) VALUES
    ('admin', 'Platform administrator with curriculum and user management rights'),
    ('learner', 'Learner account with curriculum viewing, challenges, and private notes')
ON CONFLICT (name) DO NOTHING;

-- User Roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (role_id);

-- Invitations table
CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invitations_email_lower ON invitations (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_invitations_token_hash ON invitations (token_hash);
CREATE INDEX IF NOT EXISTS idx_invitations_status ON invitations (expires_at, accepted_at, revoked_at);

-- Refresh Sessions table
CREATE TABLE IF NOT EXISTS refresh_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    token_family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_sessions_user_id ON refresh_sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_token_hash ON refresh_sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_sessions_family ON refresh_sessions (token_family_id);
