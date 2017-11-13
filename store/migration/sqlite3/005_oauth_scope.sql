-- +migrate Up

ALTER TABLE users ADD COLUMN user_scopes TEXT DEFAULT '';
