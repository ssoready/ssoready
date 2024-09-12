create type scim_request_http_method as enum ('get', 'post', 'put', 'patch', 'delete');
create type scim_request_http_status as enum ('200', '201', '204', '400', '401', '404');

create table scim_requests
(
    id                                       uuid                     not null primary key,
    scim_directory_id                        uuid                     not null references scim_directories (id),
    timestamp                                timestamptz              not null,

    http_request_url                         varchar                  not null,
    http_request_method                      scim_request_http_method not null,
    http_request_headers                     jsonb                    not null,
    http_request_body                        jsonb,

    http_response_status                     scim_request_http_status not null,
    http_response_body                       jsonb,

    error_bad_bearer_token                   boolean default false    not null,
    error_bad_username                       varchar,
    error_email_outside_organization_domains varchar
);
