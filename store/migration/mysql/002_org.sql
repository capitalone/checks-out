-- +migrate Up

ALTER TABLE repos ADD COLUMN repo_org BOOLEAN DEFAULT FALSE;

-- +migrate Down

ALTER TABLE repos DROP COLUMN repo_org;