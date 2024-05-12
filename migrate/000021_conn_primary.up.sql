alter table saml_connections
    add column is_primary bool not null default false;
