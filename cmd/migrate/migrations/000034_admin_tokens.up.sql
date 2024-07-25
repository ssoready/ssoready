create table admin_access_tokens
(
    id                    uuid        not null primary key,
    organization_id       uuid        not null references organizations (id),
    one_time_token_sha256 bytea       unique,
    access_token_sha256   bytea       unique,
    create_time           timestamptz not null,
    expire_time           timestamptz not null
);
