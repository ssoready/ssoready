alter table saml_flows
    add column error_saml_connection_not_configured boolean not null default false;
