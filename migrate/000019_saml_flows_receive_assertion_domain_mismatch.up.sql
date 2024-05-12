create type saml_flow_status as enum ('in_progress', 'failed', 'succeeded');

alter table saml_flows
    add column error_bad_issuer                         varchar,
    add column error_bad_audience                       varchar,
    add column error_bad_subject_id                     varchar,
    add column error_email_outside_organization_domains varchar,
    add column status                                   saml_flow_status;
