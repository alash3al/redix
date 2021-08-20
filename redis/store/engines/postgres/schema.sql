CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- users table
-- a user log-in using the base64(id:secret)
-- a secret is a bcrypt encrypted randomly generated string.
CREATE TABLE IF NOT EXISTS redix_users (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    secret varchar NOT NULL
);

-- databases table
-- when a user selects database no 0 we look for its id and return it.
CREATE TABLE IF NOT EXISTS redix_databases (
    id bigserial NOT NULL,

    user_id uuid NOT NULL,

    -- 0, 1, 2, 3 ,4, ... 100 ... 1000 ... etc
    alias INT DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_redix_databases_uid_alias ON redix_databases(user_id, alias);

-- redix data holds all data inserted via redis server
CREATE TABLE IF NOT EXISTS redix_data (
    id bigserial NOT NULL PRIMARY KEY,
    db_id bigint NOT NULL,

    key_name varchar NOT NULL,
    key_value jsonb default null,
    is_deleted bool default false,
    expires_at timestamp DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_redix_data_dbid_keyname ON redix_data(db_id, key_name);
