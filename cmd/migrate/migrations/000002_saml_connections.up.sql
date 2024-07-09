create table environments
(
    id           uuid not null primary key,
    redirect_url varchar
);

create table organizations
(
    id             uuid not null primary key,
    environment_id uuid not null references environments (id)
);

create table saml_connections
(
    id                   uuid not null primary key,
    organization_id      uuid not null references organizations (id),
    idp_redirect_url     varchar,
    idp_x509_certificate bytea,
    idp_entity_id        varchar
);
