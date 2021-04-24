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

    /*
        HSET N K V
    
        SET X Y

        LPUSH N v1 v2 v3

        KEYS ..
    */

    _key varchar NOT NULL,
    -- _subkey varchar DEFAULT '@',

    -- -- we can store multiple datatypes in jsonb that may be longer (in length) than varchar
    -- _value jsonb,

    PRIMARY KEY (_id)
);

-- redix_hashmaps (
--     hk
--     k
--     v
-- )

-- lists (
--     id bigserial
--     lk
--     v
-- )

-- -- kv indexes
-- CREATE UNIQUE INDEX IF NOT EXISTS idx_root_key_subkey ON redix_kv (_db, _key, _subkey);
