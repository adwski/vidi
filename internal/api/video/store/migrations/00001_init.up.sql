BEGIN TRANSACTION;

CREATE TABLE videos (
                      id VARCHAR(50) NOT NULL PRIMARY KEY,
                      user_id VARCHAR(50) NOT NULL,
                      status smallint NOT NULL,
                      name VARCHAR(200) NOT NULL,
                      location VARCHAR(100) NOT NULL DEFAULT '',
                      size bigint NOT NULL,
                      playback_meta jsonb NOT NULL DEFAULT '{}',
                      created_at timestamptz default current_timestamp,
                      CONSTRAINT id_not_empty CHECK (id != ''),
                      CONSTRAINT user_uid_not_empty CHECK (user_id != ''),
                      CONSTRAINT name_not_empty CHECK (name != ''),
                      CONSTRAINT size_positive CHECK (size > 0)
);

CREATE UNIQUE INDEX videos_id ON videos (id);
CREATE INDEX videos_user_id ON videos (user_id);
CREATE INDEX videos_status ON videos (status);
CREATE INDEX videos_created_at ON videos (created_at);

CREATE TABLE upload_parts (
                num integer NOT NULL,
                video_id VARCHAR(50) NOT NULL,
                checksum VARCHAR(32) NOT NULL,
                status smallint NOT NULL,
                size bigint NOT NULL,
                CONSTRAINT num_positive CHECK (num >=0 ),
                CONSTRAINT video_id_not_empty CHECK (video_id != ''),
                CONSTRAINT checksum_not_empty CHECK (checksum != ''),
                CONSTRAINT size_positive CHECK (size > 0),
                PRIMARY KEY (num, video_id)
);

COMMIT;
