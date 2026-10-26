-- 001_create_users.down.sql

DROP INDEX IF EXISTS ux_users_email;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_email_active;
DROP INDEX IF EXISTS idx_users_last_login_at;
DROP INDEX IF EXISTS idx_sessions_user_id_active;
DROP INDEX IF EXISTS idx_sessions_refresh_expires_at;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS tokens;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;
-- Keep pgcrypto installed by default because other schemas may depend on it.
