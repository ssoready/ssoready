create table saml_login_events
(
    id                     uuid        not null primary key,
    saml_connection_id     uuid        not null references saml_connections (id),
    access_code            uuid        not null unique,
    state                  varchar     not null,
    expire_time            timestamptz not null,
    subject_idp_id         varchar,
    subject_idp_attributes jsonb
);

create type saml_login_event_timeline_entry_type as enum ('get_redirect', 'saml_initiate', 'saml_receive_assertion', 'redeem');

create table saml_login_event_timeline_entry
(
    id                             uuid                                 not null primary key,
    saml_login_event_id            uuid                                 not null references saml_login_events (id),
    timestamp                      timestamptz                          not null,
    type                           saml_login_event_timeline_entry_type not null,
    get_redirect_url               varchar,
    saml_initiate_url              varchar,
    saml_receive_assertion_payload varchar
);
