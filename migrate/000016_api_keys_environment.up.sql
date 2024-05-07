alter table api_keys
    add column environment_id uuid not null references environments (id);
alter table api_keys drop column app_organization_id;
