-- +migrate Up

ALTER TABLE users DROP COLUMN user_email;

-- +migrate Down

ALTER TABLE users ADD COLUMN user_email VARCHAR(255) DEFAULT '';