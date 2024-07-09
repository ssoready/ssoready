drop table saml_sessions;
drop table saml_login_event_timeline_entries;
drop table saml_login_events;
drop type saml_login_event_timeline_entry_type;

create table saml_flows
(
    id                     uuid        not null primary key,
    saml_connection_id     uuid        not null references saml_connections (id),
    access_code            uuid        not null unique,
    state                  varchar     not null,
    create_time            timestamptz not null,
    expire_time            timestamptz not null,
    subject_idp_id         varchar,
    subject_idp_attributes jsonb
);

create type saml_flow_step_type as enum ('get_redirect', 'saml_initiate', 'saml_receive_assertion', 'redeem');

create table saml_flow_steps
(
    id                             uuid                 not null primary key,
    saml_flow_id                   uuid                 not null references saml_flows (id),
    timestamp                      timestamptz          not null,
    type                           saml_flow_step_type not null,
    get_redirect_url               varchar,
    saml_initiate_url              varchar,
    saml_receive_assertion_payload varchar
);
