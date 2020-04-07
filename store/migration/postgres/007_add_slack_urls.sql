-- +migrate Up

create table if not exists slack_urls(
  id BIGSERIAL PRIMARY KEY,
  host_name text not null,
  "user" text not null,
  url text not null,
  unique(host_name, "user")
);

-- +migrate Down

DROP TABLE slack_urls;
