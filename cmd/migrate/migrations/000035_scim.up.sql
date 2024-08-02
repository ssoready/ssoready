create table scim_directories
(
    id                  uuid  not null primary key,
    organization_id     uuid  not null references organizations (id),
    bearer_token_sha256 bytea not null
);

create table scim_users
(
    id                uuid    not null primary key,
    scim_directory_id uuid    not null references scim_directories (id),
    email             varchar not null,
    deleted           bool    not null,
    attributes        jsonb,

    unique (scim_directory_id, email)
);
