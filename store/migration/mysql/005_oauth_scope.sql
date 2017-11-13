-- +migrate Up

ALTER TABLE users ADD COLUMN user_scopes VARCHAR(255) DEFAULT '';

-- +migrate Down

ALTER TABLE users DROP COLUMN user_scopes;