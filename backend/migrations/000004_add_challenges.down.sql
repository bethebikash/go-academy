-- 000004_add_challenges.down.sql

DROP INDEX IF EXISTS uq_user_challenge_bookmark;
ALTER TABLE bookmarks DROP CONSTRAINT IF EXISTS fk_bookmarks_challenge;
ALTER TABLE notes DROP CONSTRAINT IF EXISTS fk_notes_challenge;

DROP TABLE IF EXISTS challenge_test_cases CASCADE;
DROP TABLE IF EXISTS challenges CASCADE;
