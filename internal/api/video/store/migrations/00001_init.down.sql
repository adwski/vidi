BEGIN TRANSACTION;

DROP INDEX videos_id;
DROP INDEX videos_user_id;
DROP INDEX videos_status;
DROP INDEX videos_created_at;

DROP TABLE videos;

COMMIT;
