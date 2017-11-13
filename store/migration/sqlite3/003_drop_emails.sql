-- +migrate Up

ALTER TABLE users RENAME TO temp_users;

CREATE TABLE IF NOT EXISTS users (
 user_id      INTEGER PRIMARY KEY AUTOINCREMENT
,user_login   TEXT
,user_token   TEXT
,user_avatar  TEXT
,user_secret  TEXT

,UNIQUE(user_login)
);

INSERT INTO users
SELECT
 user_id, user_login, user_token, user_avatar, user_secret
FROM
 temp_users;

DROP TABLE temp_users;

-- +migrate Down

ALTER TABLE users ADD COLUMN user_email TEXT DEFAULT '';