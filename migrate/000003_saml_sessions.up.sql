create table saml_sessions
(
    id                     uuid not null primary key,
    saml_connection_id     uuid not null references saml_connections (id),
    secret_access_token    uuid,
    subject_id             varchar,
    subject_idp_attributes jsonb
);
