CREATE TABLE IF NOT EXISTS repos ();

ALTER TABLE repos
ADD COLUMN IF NOT EXISTS id BIGINT PRIMARY KEY,
ADD COLUMN IF NOT EXISTS repo VARCHAR(100) NOT NULL,
ADD COLUMN IF NOT EXISTS owner VARCHAR(100) NOT NULL,
ADD COLUMN IF NOT EXISTS schedule VARCHAR(100) NOT NULL,
ADD COLUMN IF NOT EXISTS lastrun TIMESTAMP;