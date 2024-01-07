BEGIN TRANSACTION;

CREATE TABLE users (
                      id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                      uid  VARCHAR(50) NOT NULL,
                      name VARCHAR(50) NOT NULL,
                      hash VARCHAR(100) NOT NULL,
                      ts timestamp default current_timestamp,
                      CONSTRAINT uid_not_empty CHECK (uid != ''),
                      CONSTRAINT name_not_empty CHECK (name != ''),
                      CONSTRAINT hash_not_empty CHECK (hash != '')
);

CREATE UNIQUE INDEX users_uid On users (uid);
CREATE UNIQUE INDEX users_name On users (name);
CREATE INDEX users_ts ON users (ts);

COMMIT;
