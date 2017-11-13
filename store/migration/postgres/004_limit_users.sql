-- +migrate Up

CREATE TABLE IF NOT EXISTS limit_users (
  login      TEXT
  ,UNIQUE(login)
);

CREATE TABLE IF NOT EXISTS limit_orgs (
  org TEXT
  , UNIQUE(org)
);

-- +migrate Down

DROP TABLE limit_users;

DROP TABLE limit_orgs;
