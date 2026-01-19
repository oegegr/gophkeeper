
-- +migrate Down
DROP INDEX IF EXISTS idx_secrets_user_id_type;
DROP INDEX IF EXISTS idx_secrets_type;
DROP INDEX IF EXISTS idx_secrets_user_id;
DROP INDEX IF EXISTS idx_users_login;