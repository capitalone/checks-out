-- +migrate Up

CREATE TABLE IF NOT EXISTS orgs (
  org_id      INTEGER PRIMARY KEY AUTOINCREMENT,
  org_user_id INTEGER,
  org_owner   VARCHAR(255),
  org_link    VARCHAR(1024),
  org_private BOOLEAN,
  org_secret  VARCHAR(255),
  UNIQUE (org_owner)
);

CREATE INDEX IF NOT EXISTS ix_org_owner ON orgs (org_owner);
CREATE INDEX IF NOT EXISTS ix_org_user_id ON orgs (org_user_id);

-- +migrate Down

DROP TABLE orgs;
