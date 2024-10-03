alter table saml_flows
    add column error_saml_connection_not_configured boolean not null default false;
alter table saml_flows
    add column error_environment_oauth_redirect_uri_not_configured boolean not null default false;
