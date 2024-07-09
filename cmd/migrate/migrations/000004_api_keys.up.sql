create table app_organizations
(
    id uuid not null primary key
);

create table api_keys
(
    id                  uuid    not null primary key,
    app_organization_id uuid    not null references app_organizations (id),
    secret_value        varchar not null
);

alter table environments
    add column app_organization_id uuid not null references app_organizations (id);
