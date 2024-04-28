alter table app_organizations add column google_hosted_domain varchar unique;

create table app_users
(
    id                  uuid    not null primary key,
    app_organization_id uuid    not null references app_organizations (id),
    display_name        varchar not null,
    email               varchar unique
);

create table app_sessions
(
    id          uuid        not null primary key,
    app_user_id uuid        not null references app_users (id),
    create_time timestamptz not null,
    expire_time timestamptz not null,
    token       varchar     not null unique
);
