-- +migrate Up

create table if not exists slack_urls(
  id integer primary key AUTO_INCREMENT,
  host_name VARCHAR(255) not null,
  user VARCHAR(255) not null,
  url VARCHAR(1024) not null,
  unique(host_name, user)
);

-- +migrate Down

DROP TABLE slack_urls;
