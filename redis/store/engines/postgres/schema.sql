CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- users table
-- a user log-in using the base64(id:secret)
-- a secret is a bcrypt encrypted randomly generated string.
CREATE TABLE IF NOT EXISTS redix_users (
    id uuid DEFAULT uuid_generate_v4(),
    secret varchar NOT NULL,

    PRIMARY KEY(id)
);

-- databases table
-- when a user selects database no 0 we look for its id and return it.
CREATE TABLE IF NOT EXISTS redix_databases (
    id uuid DEFAULT uuid_generate_v4(),

    -- TODO:// move to multiple users per db
    user_id uuid NOT Null,
    
    name varchar(16) NOT Null,

    PRIMARY KEY(id)
);

-- database table indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_databases_userid_name ON redix_databases (user_id, name);

-- root table
-- this is the root table that holds the keys as well its meta info
-- read more about redis datatypes from here: https://redis.io/topics/data-types
CREATE TABLE IF NOT EXISTS redix_meta (
    _id bigserial NOT NULL,
    _db uuid NOT NULL,
    _type varchar(20) DEFAULT 'str',
    _key varchar NOT NULL,

    PRIMARY KEY (_id)
);

CREATE TABLE IF NOT EXISTS redix_strings (
    _key_id bigserial NOT NULL,
    _value jsonb not null,

    PRIMARY KEY(_key_id)
);

CREATE TABLE IF NOT EXISTS redix_numbers (
    _key_id bigserial NOT NULL,
    _value numeric DEFAULT 0.0,

    PRIMARY KEY(_key_id)
);
