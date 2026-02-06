DROP INDEX IF EXISTS idx_secrets_owner_type;
DROP INDEX IF EXISTS idx_secrets_owner_updated;

DROP TABLE IF EXISTS secrets;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS citext;
