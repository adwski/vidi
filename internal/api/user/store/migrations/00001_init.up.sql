BEGIN TRANSACTION;

CREATE TABLE users (
                      id VARCHAR(50) NOT NULL PRIMARY KEY,
                      name VARCHAR(50) NOT NULL,
                      hash VARCHAR(100) NOT NULL,
                      created_at timestamptz default current_timestamp,
                      CONSTRAINT id_not_empty CHECK (id != ''),
                      CONSTRAINT name_not_empty CHECK (name != ''),
                      CONSTRAINT hash_not_empty CHECK (hash != '')

);

CREATE UNIQUE INDEX users_id ON users (id);
CREATE UNIQUE INDEX users_name ON users (name);
CREATE INDEX users_created_at ON users (created_at);

COMMIT;
