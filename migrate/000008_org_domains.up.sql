create table organization_domains (
    id uuid not null primary key,
    organization_id uuid not null references organizations(id),
    domain varchar not null
);
