-- +migrate Up

CREATE TABLE IF NOT EXISTS orgs (
   org_id       INTEGER PRIMARY KEY AUTO_INCREMENT
  ,org_user_id  INTEGER
  ,org_owner    VARCHAR(255)
  ,org_link     VARCHAR(1024)
  ,org_private  BOOLEAN
  ,org_secret   VARCHAR(255)
  ,UNIQUE(org_owner)
);

ALTER TABLE orgs ADD INDEX (org_owner);
ALTER TABLE orgs ADD INDEX (org_user_id);

-- +migrate Down

DROP TABLE orgs;
