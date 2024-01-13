BEGIN TRANSACTION;

CREATE TABLE videos (
                      id VARCHAR(50) NOT NULL PRIMARY KEY,
                      user_id VARCHAR(50) NOT NULL,
                      status smallint NOT NULL,
                      location VARCHAR(100) NOT NULL DEFAULT '',
                      created_at timestamptz default current_timestamp,
                      CONSTRAINT id_not_empty CHECK (id != ''),
                      CONSTRAINT user_uid_not_empty CHECK (user_id != '')
);

CREATE UNIQUE INDEX videos_id On videos (id);
CREATE INDEX videos_user_id On videos (user_id);
CREATE INDEX videos_status ON videos (status);
CREATE INDEX videos_created_at ON videos (created_at);

COMMIT;
