-- 001_create_users.down.sql

DROP INDEX IF EXISTS ux_users_email;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;

-- Keep pgcrypto installed by default because other schemas may depend on it.
