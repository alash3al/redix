create table if not exists redix_v3_data (
    _id bigserial primary key,
    _key text,
    _value bytea,
    _created_at timestamp default current_timestamp,
    _updated_at timestamp default current_timestamp
);