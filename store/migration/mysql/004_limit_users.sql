-- +migrate Up

CREATE TABLE IF NOT EXISTS limit_users (
  login      VARCHAR(255) NOT NULL
,UNIQUE(login)
);

CREATE TABLE IF NOT EXISTS limit_orgs (
  org VARCHAR(255) NOT NULL
  , UNIQUE(org)
);

-- +migrate Down

DROP TABLE limit_users;

DROP TABLE limit_orgs;
