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
    user_id uuid NOT Null,
    name varchar(16) NOT Null,

    PRIMARY KEY(id)
);

-- database table indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_databases_userid_name ON redix_databases (user_id, name);