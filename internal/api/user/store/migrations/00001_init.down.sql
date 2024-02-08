BEGIN TRANSACTION;

DROP INDEX users_id;
DROP INDEX users_name;
DROP INDEX users_created_at;

DROP TABLE users;

COMMIT;
