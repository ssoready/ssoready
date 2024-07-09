drop table saml_flow_steps;
drop type saml_flow_step_type;

alter table saml_flows
    add column update_time            timestamptz not null,
    add column auth_redirect_url      varchar,
    add column get_redirect_time      timestamptz,
    add column initiate_request       varchar,
    add column initiate_time          timestamptz,
    add column assertion              varchar,
    add column app_redirect_url       varchar,
    add column receive_assertion_time timestamptz,
    add column redeem_time            timestamptz;
