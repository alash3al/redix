CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- users table
-- a user log-in using the base64(id:secret)
-- a secret is a bcrypt encrypted randomly generated string.
CREATE TABLE IF NOT EXISTS redix_users (
    id uuid DEFAULT uuid_generate_v4() PRIMARY KEY,

    password varchar NOT NULL
);

-- databases table
-- when a user selects database no 0 we look for its id and return it.
CREATE TABLE IF NOT EXISTS redix_databases (
    id bigserial NOT NULL,

    user_id uuid NOT NULL,

    -- 0, 1, 2, 3 ,4, ... 100 ... 1000 ... etc
    alias INT DEFAULT 0,

    UNIQUE INDEX (user_id, dbkey);
);

-- database table indexes
-- TODO

-- root table
-- this is the root table that holds the keys as well its meta info
-- read more about redis datatypes from here: https://redis.io/topics/data-types
CREATE TABLE IF NOT EXISTS redix_keys_meta (
    id bigserial NOT NULL,
    db_id bigint NOT NULL,

    key_type varchar(20) DEFAULT 'str',
    key_name varchar NOT NULL,
    
    expires_at timestamp DEFAULT NULL,

    PRIMARY KEY (_id)
);

CREATE TABLE IF NOT EXISTS redix_values_string (
    key_id bigserial NOT NULL,
    
    value bytea not null,

    PRIMARY KEY(key_id)
);

CREATE TABLE IF NOT EXISTS redix_values_number (
    key_id bigserial NOT NULL,
    value numeric DEFAULT 0.0,

    PRIMARY KEY(key_id)
);
