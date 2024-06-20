create table saml_oauth_clients
(
    id                   uuid  not null primary key,
    environment_id       uuid  not null references environments (id),
    client_secret_sha256 bytea not null unique
);
