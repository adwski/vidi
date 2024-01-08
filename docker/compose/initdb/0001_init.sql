CREATE USER userapi
    PASSWORD 'userapi';

CREATE DATABASE userapi
    OWNER userapi
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';

CREATE USER videoapi
    PASSWORD 'videoapi';

CREATE DATABASE videoapi
    OWNER videoapi
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';
