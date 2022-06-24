CREATE TABLE IF NOT EXISTS repos (
  id bigint PRIMARY KEY,
  repo varchar(100),
  owner varchar(100),
  schedule varchar(100)
  lastrun timestamp,
);