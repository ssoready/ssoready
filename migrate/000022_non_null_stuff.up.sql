alter table app_users
    alter column email set not null;
alter table saml_connections alter column sp_entity_id set not null;
alter table saml_connections
    add column sp_acs_url varchar not null;
alter table saml_flows
    rename column subject_idp_id to email;
alter table saml_flows
    alter column status set not null;
